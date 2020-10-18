package ytcompare

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"

	"github.com/tencentyun/cos-go-sdk-v5"
)

//Compare compare struct
type Compare struct {
	httpCli   *http.Client
	dbCli     *mongo.Client
	cosCli    *cos.Client
	dbName    string
	SyncURLs  []string
	StartTime int
	TimeRange int
	WaitTime  int
	SkipTime  int
}

//New create a new Compare instance
func New(ctx context.Context, config *Config) (*Compare, error) {
	entry := log.WithFields(log.Fields{Function: "New"})
	dbClient, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoDBURL))
	if err != nil {
		entry.WithError(err).Errorf("creating mongo DB client failed: %s", config.MongoDBURL)
		return nil, err
	}
	COSURL := fmt.Sprintf("%s://%s.%s", config.COS.Schema, config.COS.BucketName, config.COS.Domain)
	u, err := url.Parse(COSURL)
	if err != nil {
		entry.WithError(err).Errorf("parse COS URL failed: %s", COSURL)
		return nil, err
	}
	cosClient := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.COS.SecretID,
			SecretKey: config.COS.SecretKey,
		},
	})
	return &Compare{httpCli: &http.Client{}, dbCli: dbClient, cosCli: cosClient, dbName: config.DBName, SyncURLs: config.AllSyncURLs, StartTime: config.StartTime, TimeRange: config.TimeRange, WaitTime: config.WaitTime, SkipTime: config.SkipTime}, nil
}

//Start start compare service
func (compare *Compare) Start(ctx context.Context) {
	entry := log.WithFields(log.Fields{Function: "Start"})
	urls := compare.SyncURLs
	snCount := len(urls)
	checkPointTab := compare.dbCli.Database(compare.dbName).Collection(CheckPointTab)
	entry.Info("compare service starting")
	store := NewStore()
	for {
		var checkPointOld *CheckPoint
		checkPoint := new(CheckPoint)
		err := checkPointTab.FindOne(ctx, bson.M{"_id": 1}).Decode(checkPoint)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				checkPoint = &CheckPoint{ID: 1, Start: int64(compare.StartTime), Range: int64(compare.TimeRange), Timestamp: time.Now().Unix()}
				entry.Debugf("no checkpoint record")
			} else {
				entry.WithError(err).Error("fetch checkpoint record")
				time.Sleep(time.Duration(compare.WaitTime) * time.Second)
				continue
			}
		} else {
			checkPointOld = new(CheckPoint)
			checkPointOld.ID = checkPoint.ID
			checkPointOld.Start = checkPoint.Start
			checkPointOld.Range = checkPoint.Range
			checkPointOld.Timestamp = checkPoint.Timestamp
			checkPoint.Start = checkPointOld.Start + checkPointOld.Range
			checkPoint.Range = int64(compare.TimeRange)
			checkPoint.Timestamp = time.Now().Unix()
			entry.Debugf("old checkpoint: %+v", checkPointOld)
			entry.Debugf("new checkpoint: %+v", checkPoint)
		}

		if checkPoint.Start+checkPoint.Range > time.Now().Unix()-int64(compare.SkipTime) {
			entry.Debugf("time invalid: %d", checkPoint.Start+checkPoint.Range)
			time.Sleep(time.Duration(compare.WaitTime) * time.Second)
			continue
		}

		entry.Infof("fetching shards from %d to %d", checkPoint.Start, checkPoint.Start+checkPoint.Range)
		var wg sync.WaitGroup
		wg.Add(snCount)
		var innerErr *error
		for i := 0; i < snCount; i++ {
			snID := int32(i)
			go func() {
				defer wg.Done()
				entry := log.WithFields(log.Fields{Function: "Start", SNID: snID})
				entry.Debugf("starting fetching shards in SN%d from %d to %d", snID, checkPoint.Start, checkPoint.Start+checkPoint.Range)
				shards, err := GetCompareShards(compare.httpCli, urls[snID], checkPoint.Start, checkPoint.Start+checkPoint.Range)
				if err != nil {
					innerErr = &err
					entry.WithError(err).Error("fetch compare shards")
					return
				}
				for _, shard := range shards {
					store.Add(shard.NodeID, shard.VHF)
				}
			}()
		}
		wg.Wait()
		if innerErr != nil {
			store.Clear()
			time.Sleep(time.Duration(compare.WaitTime) * time.Second)
			entry.Warnf("retry fetching shards from %d to %d", checkPoint.Start, checkPoint.Start+checkPoint.Range)
			continue
		}
		entry.Infof("uploading compare data from %d to %d", checkPoint.Start, checkPoint.Start+checkPoint.Range)
		var wg2 sync.WaitGroup
		wg2.Add(len(store.Items))
		for id := range store.Items {
			nid := id
			go func() {
				defer wg2.Done()
				entry := log.WithFields(log.Fields{Function: "Start", MinerID: nid})
				if len(store.Items[nid]) == 0 {
					entry.Debugf("no compare data for uploading from %d to %d", checkPoint.Start, checkPoint.Start+checkPoint.Range)
					return
				}
				entry.Debugf("starting generating compare data from %d to %d", checkPoint.Start, checkPoint.Start+checkPoint.Range)
				data, err := store.GenerateData(nid)
				if err != nil {
					innerErr = &err
					entry.WithError(err).Errorf("generating compare data from %d to %d", checkPoint.Start, checkPoint.Start+checkPoint.Range)
					return
				}
				err = compare.UploadData(ctx, nid, data, checkPoint.Start, checkPoint.Range)
				if err != nil {
					innerErr = &err
					entry.WithError(err).Errorf("uploading compare data from %d to %d", checkPoint.Start, checkPoint.Start+checkPoint.Range)
					return
				}
			}()
		}
		wg2.Wait()
		if innerErr != nil {
			store.Clear()
			time.Sleep(time.Duration(compare.WaitTime) * time.Second)
			entry.Warnf("retry uploading shards from %d to %d", checkPoint.Start, checkPoint.Start+checkPoint.Range)
			continue
		}
		store.Clear()
		if checkPointOld == nil {
			_, err := checkPointTab.InsertOne(ctx, checkPoint)
			if err != nil {
				entry.WithError(err).Errorf("insert checkpoint record: %+v", checkPoint)
			} else {
				entry.Debugf("insert checkpoint record: %+v", checkPoint)
			}
		} else {
			_, err := checkPointTab.UpdateOne(ctx, bson.M{"_id": checkPoint.ID}, bson.M{"$set": bson.M{"start": checkPoint.Start, "range": checkPoint.Range, "timestamp": checkPoint.Timestamp}})
			if err != nil {
				entry.WithError(err).Errorf("update checkpoint record: %+v", checkPoint)
			} else {
				entry.Debugf("update checkpoint record: %+v", checkPoint)
			}
		}
	}
}

//UploadData update compare data to tencent COS
func (compare *Compare) UploadData(ctx context.Context, nodeID int32, data bytes.Buffer, start int64, timeRange int64) error {
	entry := log.WithFields(log.Fields{Function: "UploadData", MinerID: nodeID})
	cursorTab := compare.dbCli.Database(compare.dbName).Collection(CursorTab)
	var cursorOld *Cursor
	cursor := new(Cursor)
	err := cursorTab.FindOne(ctx, bson.M{"_id": nodeID}).Decode(cursor)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			cursor = &Cursor{ID: nodeID, From: start, Range: timeRange, FileFrom: 0, Timestamp: time.Now().Unix()}
			entry.Debugf("no cursor record")
		} else {
			entry.WithError(err).Error("fetch cursor record")
			return err
		}
	} else {
		cursorOld = new(Cursor)
		cursorOld.ID = cursor.ID
		cursorOld.From = cursor.From
		cursorOld.Range = cursor.Range
		cursorOld.FileFrom = cursor.FileFrom
		cursorOld.Timestamp = cursor.Timestamp
		cursor.From = start
		cursor.Range = timeRange
		cursor.FileFrom = start
		cursor.Timestamp = time.Now().Unix()
	}
	_, err = compare.cosCli.Object.Put(ctx, fmt.Sprintf("%d_%d", nodeID, cursor.FileFrom), bytes.NewReader(data.Bytes()), nil)
	if err != nil {
		entry.WithError(err).Errorf("uploading data to COS from %d, range %d", cursor.From, cursor.Range)
		return err
	}
	entry.Debugf("uploading data to COS from %d, range %d", cursor.From, cursor.Range)
	if cursorOld != nil {
		opt := &cos.ObjectPutTaggingOptions{
			TagSet: []cos.ObjectTaggingTag{
				{
					Key:   "next",
					Value: fmt.Sprintf("%d_%d", nodeID, cursor.FileFrom),
				},
				{
					Key:   "range",
					Value: fmt.Sprintf("%d", cursorOld.Range),
				},
			},
		}
		_, err := compare.cosCli.Object.PutTagging(ctx, fmt.Sprintf("%d_%d", nodeID, cursorOld.FileFrom), opt)
		if err != nil {
			entry.WithError(err).Errorf("tagging data of %s failed", fmt.Sprintf("%d_%d", nodeID, cursorOld.From))
			return err
		}
		_, err = cursorTab.UpdateOne(ctx, bson.M{"_id": cursor.ID}, bson.M{"$set": bson.M{"from": cursor.From, "range": cursor.Range, "fileFrom": cursor.FileFrom, "timestamp": cursor.Timestamp}})
		if err != nil {
			entry.WithError(err).Errorf("update cursor record: %+v", cursor)
		} else {
			entry.Debugf("update cursor record: %+v", cursor)
		}
		return nil
	}
	_, err = cursorTab.InsertOne(ctx, cursor)
	if err != nil {
		entry.WithError(err).Errorf("insert cursor record: %+v", cursor)
	} else {
		entry.Debugf("insert cursor record: %+v", cursor)
	}
	return nil
}

//GetCompareShards find shards data for comparing
func GetCompareShards(httpCli *http.Client, url string, from int64, to int64) ([]*Shard, error) {
	entry := log.WithFields(log.Fields{Function: "GetCompareShards"})
	fullURL := fmt.Sprintf("%s/sync/GetStoredShards?from=%d&to=%d", url, from, to)
	entry.Debugf("fetching compare data by URL: %s", fullURL)
	request, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		entry.WithError(err).Errorf("create request failed: %s", fullURL)
		return nil, err
	}
	request.Header.Add("Accept-Encoding", "gzip")
	resp, err := httpCli.Do(request)
	if err != nil {
		entry.WithError(err).Errorf("get compare data failed: %s", fullURL)
		return nil, err
	}
	defer resp.Body.Close()
	reader := io.Reader(resp.Body)
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gbuf, err := gzip.NewReader(reader)
		if err != nil {
			entry.WithError(err).Errorf("decompress response body: %s", fullURL)
			return nil, err
		}
		reader = io.Reader(gbuf)
		defer gbuf.Close()
	}
	response := make([]*Shard, 0)
	err = json.NewDecoder(reader).Decode(&response)
	if err != nil {
		entry.WithError(err).Errorf("decode compare data failed: %s", fullURL)
		return nil, err
	}
	return response, nil
}

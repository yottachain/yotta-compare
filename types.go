package ytcompare

const (
	//Function tag
	Function = "function"
	//SNID tag
	SNID = "snID"
	//MinerID tag
	MinerID = "minerID"
	//BlockID tag
	BlockID = "blockID"
	//ShardID tag
	ShardID = "shardID"
)

const (
	//CheckPointTab CheckPoint table
	CheckPointTab = "CheckPoint"
	//CursorTab cursor table
	CursorTab = "cursor"
)

//Shard struct
type Shard struct {
	ID      int64  `bson:"_id" json:"_id"`
	NodeID  int32  `bson:"nodeId" json:"nid"`
	VHF     []byte `bson:"VHF" json:"VHF"`
	BlockID int64  `bson:"blockid" json:"bid"`
}

//CheckPoint struct
type CheckPoint struct {
	ID        int32 `bson:"_id"`
	Start     int64 `bson:"start"`
	Range     int64 `bson:"range"`
	Timestamp int64 `bson:"timestamp"`
}

//Cursor struct
type Cursor struct {
	ID        int32 `bson:"_id"`
	From      int64 `bson:"from"`
	Range     int64 `bson:"range"`
	FileFrom  int64 `bson:"fileFrom"`
	Timestamp int64 `bson:"timestamp"`
}

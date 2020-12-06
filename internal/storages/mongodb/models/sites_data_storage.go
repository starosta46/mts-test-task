package models

//SiteData is a struct to save data in mongo
type SiteData struct {
	ID         string `bson:"_id"`
	URL        string `bson:"url"`
	Data       string `bson:"data"`
	CreateDate int    `bson:"create_date"`
}

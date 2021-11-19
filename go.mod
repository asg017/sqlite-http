module github.com/asg017/sqlite-http

go 1.17

require (
	github.com/augmentable-dev/vtab v0.0.0-20210818144031-5c7659b723dd
	go.riyazali.net/sqlite v0.0.0-20210707161919-414349b4032a
)

require github.com/mattn/go-pointer v0.0.1 // indirect

replace go.riyazali.net/sqlite => github.com/asg017/sqlite v0.0.0-20211113023900-40ac9580ba56

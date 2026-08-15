package models

import "time"

type Notification struct {
	ID			int64
	UserID		int64
	Type		string 
	Title		string 
	Content		string 
	IsRead		bool 
	CreatedAt 	time.Time
}
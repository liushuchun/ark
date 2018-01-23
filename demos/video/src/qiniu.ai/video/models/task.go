package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	Task *_Task

	taskCollection = "task"
	taskIndexes    = []mgo.Index{
		{
			Key:    []string{"_id"},
			Unique: true,
		},
	}
)

type TaskModel struct {
	Id          bson.ObjectId `bson:"_id" json:"id"`
	Src         string        `bson:"src" json:"src"`
	Name        string        `bson:"name" json:"name"`
	CreateTime  time.Time     `bson:"create_time" json:"create_time"`
	DoneTime    time.Time     `bson:"done_time" json:"done_time"`
	Choice      string        `bson:"choice" json:"choice"`
	TotalSecond int           `bson:"total_second" json:"total_second"`
	Results     []ResultBody  `bson:"results" json:"results"`
	Status      TaskStatus    `bson:"status" json:"status"`
	isNewRecord bool          `bson:"-" json:"-"`
}

func NewTaskModel(src string, name string, choice string) *TaskModel {
	return &TaskModel{
		Src:         src,
		Id:          bson.NewObjectId(),
		Name:        name,
		Choice:      choice,
		CreateTime:  time.Now().UTC(),
		Status:      TaskstatusPorcessing,
		isNewRecord: true,
	}
}

func (task *TaskModel) IsNewRecord() bool {
	return task.isNewRecord
}

func (task *TaskModel) Save() (err error) {
	if !task.Id.Valid() || !task.Status.IsValid() {
		err = ErrInvalidId
		return
	}

	Task.Query(func(c *mgo.Collection) {
		if task.IsNewRecord() {
			err = c.Insert(task)

			if err == nil {
				task.isNewRecord = false
			}
		} else {
			migrations := bson.M{
				"done_time":    task.DoneTime,
				"total_second": task.TotalSecond,
				"results":      task.Results,
				"status":       task.Status,
			}

			err = c.UpdateId(task.Id, bson.M{
				"$set": migrations,
			})
		}
	})
	return
}

func (_ *_Task) UpdateResultById(id string, results []ResultBody) (err error) {
	t := time.Now().UTC()
	migrations := bson.M{
		"status":    TaskStatusDone,
		"results":   results,
		"done_time": t,
	}

	Task.Query(func(c *mgo.Collection) {
		err = c.UpdateId(bson.ObjectIdHex(id), migrations)
	})
	return
}

func (_ *_Task) Find(id string) (model *TaskModel, err error) {
	if !bson.IsObjectIdHex(id) {
		return nil, ErrInvalidId
	}

	Task.Query(func(c *mgo.Collection) {
		err = c.FindId(bson.ObjectIdHex(id)).One(&model)
	})

	return
}

type _Task struct {
}

func (_ *_Task) Query(query func(c *mgo.Collection)) {
	mongo.Query(taskCollection, taskIndexes, query)
}

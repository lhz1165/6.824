package mr

import (
	"log"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

type TaskType int

type Master struct {
	mu sync.Mutex
	// Your definitions here.
	TaskType        TaskType
	nMapTaskNums    int
	nReduceTaskNums int
	//任务的完成情况
	mapTaskDone       []bool
	mapTaskStartTime  []time.Time
	mapTaskFinishTime []time.Time

	reduceTaskDone       []bool
	reduceTaskStartTime  []time.Time
	reduceTaskFinishTime []time.Time
	//读写文件地址
	mapSeq int
	isDone bool
}
type AskTaskReq struct {
}

type AskTaskResp struct {
	TaskType TaskType
	index    int
}

type FinishTaskReq struct {
	TaskType TaskType
	index    int
}
type FinishTaskResp struct {
}

// Your code here -- RPC handlers for the worker to call.
func (m *Master) AskTaskHandler(req AskTaskReq, resp AskTaskResp) {
	//map
	isDone := false
	for {
		for index, done := range m.mapTaskDone {
			//没有做的任务派发出去
			if !done {
				resp.TaskType = 1
				m.mapTaskDone[index] = true
				m.mapTaskStartTime[index] = time.Now()
				return
			}
		}
		//如果所有任务都派发出去了等待他们都做完
		//失败的 超时的重新发布
		if !isDone {

		} else {
			//所有的map完成 派发reduce
			break
		}
	}

	//reduce
	for {
		for index, done := range m.reduceTaskDone {
			//没有做的任务派发出去
			if !done {
				resp.TaskType = 1
				m.reduceTaskDone[index] = true
				m.reduceTaskStartTime[index] = time.Now()
				return
			}
		}
		//如果所有任务都派发出去了等待他们都做完
		//失败的 超时的重新发布
		if !isDone {

		} else {
			//所有的map完成 派发reduce
			break
		}
	}

}

func (m *Master) FinishTaskHandler(req FinishTaskReq, resp FinishTaskResp) {
	taskType := req.TaskType
	index := req.index
	//完成时间
	if taskType == 1 {
		m.mapTaskFinishTime[index] = time.Now()
	} else {
		m.reduceTaskFinishTime[index] = time.Now()
	}
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false

	// Your code here.

	return ret
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// Your code here.
	m.nMapTaskNums = len(files)
	m.nReduceTaskNums = nReduce
	m.mapTaskDone = make([]bool, m.nMapTaskNums)
	m.mapTaskStartTime = make([]time.Time, m.nMapTaskNums)

	m.reduceTaskDone = make([]bool, m.nReduceTaskNums)
	m.reduceTaskStartTime = make([]time.Time, m.nReduceTaskNums)
	m.mapSeq = 0
	m.isDone = false
	m.server()
	return &m
}

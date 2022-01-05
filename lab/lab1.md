# lab1

## master数据结构

```
type Master struct {
   mu sync.Mutex
   // Your definitions here.
   TaskType        TaskType //1代表map 2代表reduce
   nMapTaskNums    int //map任务个数，几个文件几个map
   nReduceTaskNums int   //rendu任务个数，自定义输入 实验为10
   
   //任务的完成情况
   mapTaskDone       []bool //派发的map完成情况，大小为nMapTaskNums
   mapTaskStartTime  []time.Time //派发的map开始时间，大小为nMapTaskNums
   mapTaskFinishTime []time.Time /派发的map完成情况，大小为nMapTaskNums
	//同上
   reduceTaskDone       []bool
   reduceTaskStartTime  []time.Time
   reduceTaskFinishTime []time.Time
   //是否完成，当所有map和redue任务完成了就设置为true
   isDone bool
   //需要处理的文件
   files  []string
}
```



## RPC接口

一个是worker请求任务的rpc，另一个是worker完成任务的rpc

请求任务(包括map和reduce)

```
func (m *Master) AskTaskHandler(req *AskTaskReq, resp *AskTaskResp) error 
```

先派发map任务，所有map任务完成好了再派发reduce任务

```go
func (m *Master) AskTaskHandler(req *AskTaskReq, resp *AskTaskResp) error {
   m.mu.Lock()
   defer m.mu.Unlock()
   if m.isDone {
      fmt.Println("all task done.....")
      resp.TaskType = 3
      return nil
   }
   resp.NReduce = m.nReduceTaskNums
   resp.NMap = m.nMapTaskNums
   //map
   for {
      fmt.Println("start choose a map task")
      mapDone := true
      for index, done := range m.mapTaskDone {
         //没有上报完成的任务
         if !done {
            fmt.Println("choose unfinished map task...")
            //【未开始】 或者 【超时】的任务派发出去
            if m.mapTaskStartTime[index].IsZero() ||
               time.Since(m.mapTaskStartTime[index]).Seconds() > 10 {
               resp.TaskType = 1
               resp.Index = index
               resp.TaskFileName = m.files[index]
               m.mapTaskStartTime[index] = time.Now()
               fmt.Printf("【choose unfinished map task】 task index is %d\n", index)
               return nil
            } else {
               //如果有任务还在运行中
               mapDone = false
               fmt.Println("【choose unfinished map task】 all task running, wait and choose again....")
            }
         }
      }
      //如果任务都没超时，等待，在看看是否需要继续分配
      if !mapDone {
         //sleep
         time.Sleep(time.Second * 1)
         fmt.Println("map task running wait....")
      } else {
         fmt.Println("map task done!!!!")
         break
      }
   }
   //分配reduce任务
   for {
      fmt.Println("start choose a reduce task")
      reduceDone := true
      for index, done := range m.reduceTaskDone {
         //没有上报完成的任务
         if !done {
            fmt.Println("choose unfinished reduce task...")
            //【未开始】 或者 【超时】的任务派发出去
            if m.reduceTaskStartTime[index].IsZero() ||
               time.Since(m.reduceTaskStartTime[index]).Seconds() > 10 {
               resp.TaskType = 2
               resp.Index = index
               m.reduceTaskStartTime[index] = time.Now()
               fmt.Printf("【choose unfinished reduce task】 task index is %d\n", index)
               return nil
            } else {
               //如果有任务还在运行中
               reduceDone = false
               fmt.Println("【choose unfinished reduce task】 all task running, wait and choose again....")
            }
         }
      }
      //如果任务都没超时，等待，在看看是否需要继续分配
      if !reduceDone {
         //sleep
         time.Sleep(time.Second * 1)
         fmt.Println("reduce task running wait....")
      } else {
         fmt.Println("reduce task done!!!!")
         break
      }
   }
   //reduce也分配完了
   fmt.Println("all task done!!!!!")
   resp.TaskType = 3
   m.isDone = true
   return nil
}
```



完成任务(标记任务的完成和完成时间)

```
func (m *Master) FinishTaskHandler(req *FinishTaskReq, resp *FinishTaskResp) error 
```



```go
func (m *Master) FinishTaskHandler(req *FinishTaskReq, resp *FinishTaskResp) error {
   m.mu.Lock()
   defer m.mu.Unlock()
   taskType := req.TaskType
   index := req.Index
   //完成时间
   if taskType == 1 {
      fmt.Printf("【finished map task】 task index is %d\n", index)
      m.mapTaskFinishTime[index] = time.Now()
      m.mapTaskDone[index] = true
   } else if taskType == 2 {
      fmt.Printf("【finished reduce task】 task index is %d\n", index)
      m.reduceTaskFinishTime[index] = time.Now()
      m.reduceTaskDone[index] = true
   }
   return nil
}
```



## Worker

核心就是轮询，不断请求任务，每次做完任务就发送个rpc报告完成，直到master告诉worker所有任务做完了就退出程序，

```go
func Worker(mapf func(string, string) []KeyValue,
   reducef func(string, []string) string) {
   for {
      req := AskTaskReq{}
      resp := AskTaskResp{}
      call("Master.AskTaskHandler", &req, &resp)
      switch resp.TaskType {
      case TaskType(1):
         doMapTask(resp.TaskFileName, resp.NReduce, resp.Index, mapf)
         time.Sleep(time.Second * 1)
      case TaskType(2):
         doReduceTask(resp.NMap, resp.Index, reducef)
         time.Sleep(time.Second * 1)
      case TaskType(3):
         fmt.Println("all over")
         os.Exit(0)
      default:
         fmt.Println("error task type")
      }
      finreq := FinishTaskReq{TaskType: resp.TaskType, Index: resp.Index}
      finresp := FinishTaskResp{}
      call("Master.FinishTaskHandler", &finreq, &finresp)
   }

}
```
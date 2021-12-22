# 6.824
mit lab
cd src/main
go build -buildmode=plugin ../mrapps/wc.go
rm mr-out*
go run mrsequential.go wc.so pg*.txt
more mr-out-0

提示:
- 开始工作的一种方式是修改mr/worker.go文件的Worker()去给master发送一个rpc来请求任务，之后修改master去返回一个未开始map 任务的文件名。然后修改worker 去读文件并且调用map方法就像mrsequential.go
- 这个应用的map和reduce方法在运行时被加载到 Go plugin package，就是.so文件
- 如果你修改了mr/目录下的文件,你应该需要re-build，go build -buildmode=plugin ../mrapps/wc.go
- 该lab依赖于worker的共享文件系统，当所有worker都跑在相同机器上时它们是很简单的，但是如果worker跑在不同的机器则需要一个像gfs的分布式的文件系统
- 按照管理对于中间(intermediate)文件合理的命名是mr-X-Y，x代表map 任务的编号，t代表reduce的编号
- worker的map任务代码需要某种方式去在文件里存储中间(intermediate)k/v对，在reduce任务期间能正确的读回。一种可能的方式是使用go的encoding/json包。把k/v对写道json文件中
```
 enc := json.NewEncoder(file)
  for _, kv := ... {
    err := enc.Encode(&kv)
```
然后读回这样的文件
```
  dec := json.NewDecoder(file)
  for {
    var kv KeyValue
    if err := dec.Decode(&kv); err != nil {
      break
    }
    kva = append(kva, kv)
  }
```
- worker关于map部分的可以使用ihash(key)方法，去给给定的key选择reduce任务
- 你可以从mrsequential.go偷一点代码，去实现读map输入的文件，去实现在map和reduce中间排序 中间(intermedate )k/v对，和去实现保存Reduce的输出到文件
- master作为一个rpc的服务器，将实现并行；不要忘了锁一下data
- 使用go race探测， go build -race，go run -race. test-mr.sh有一个comment去向你展示如何为测试开启race探测器
- worker有时候需要等待，举个例子：reduce直到最近一个map完成了才开始。一种可能是worker定期向master请求任务，在每个请求之间睡眠time.Sleep()。另一个可能是，在master里相关的rpc处理器(handler)有个等待的循环，time.Sleep() 或者 sync.Cond。Go 为每个rpc在它自己的线程里启动handler，因此实际上当一个handler处于等待，并不会阻止master处理别的rpc
- master不能区分worker是否宕机，worker有可能活着但是因为别的原因停滞了，或者处理太慢了，最好的方式是等待一段时间然后直接放弃。在这个实验室中，让master等待十秒钟；之后，master应该假定该worker已经死亡（当然，它可能没有死亡）。
- 为了测试崩溃后恢复，你可以使用mrapps/crash.go插件，它可以在map和reduce方法中随机的退出
- 为了确保在崩溃的情况下没有人观察到部分写入的文件,MapReduce paper提到使用一个临时文件并且原子性的重命名它一旦完成了写入。你可以使用ioutil.TempFile去创建一个临时文件并且os.Rename去原子的重命名
- test-mr.sh启动整个程序在mr-tmp子目录，所以如果有一些错误或者 你想看中间(intermediate )文件或输出文件，那就看那里
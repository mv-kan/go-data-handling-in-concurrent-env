# Go Data Handling in Concurrent Environment
Problems that are solved:
- In concurrent env data can be easily out of date (race condition) and using mutex is kinda tricky so this is cheatsheet for myself. 
- This architecture gives ability to handle data "on update"
- In general this repo gives a direction and an architecture that you can use as inspiration

## How to use Data Store

## Data Store
Terminology 
- `owner` is goroutine that is the only owner of a data store that have ultimate authority over the data. It can lock, unlock data, send notifications and change the data directly
- `consumer` is goroutine that "consumes" data store. Consumer interfaces are (in `pkg/store/types.go`) `DataStoreGetter`, `DataStoreSetter` but why not `DataStoreNotifier`? It is better to create new channels and `Fanout[T]()` data store notifier channel. 
- `fire up function` is a function that returns necessary objects/interfaces/channels and fires up goroutine that looks like 
```
go func(){
for {
    select{
        ...
    }
}
}()
```
You should see examples to find out how this all glued up together. But generally 
- owner fire up function MUST NOT return `DataStoreOwner` interface, but can `DataStore`
- owner fire up fn MUST return interface that has `Notifier` because we need to create `Fanout[T]()` for notifier
 
## Car Example 

We have `carDevice` object that has methods like `IsConnected()` `GetSpeed()` `SetSpeed(float64) error` and `GetLocation()`. Let's imagine that these methods come from some API library of car manufacturer, Ok? Good. 

And in this example we need to 
1. Print out `GetSpeed()` and `GetLocation()` every couple of seconds
2. Save `GetSpeed()` and `GetLocation()` in file `log.txt` with date like this `<date>: <speed>, <location>`
3. Increase speed by 1 using `SetSpeed(float64)` every 5 seconds
4. Have ability to kill every go routine at any point of time 

### Let's talk about architecture and how I solve this

`RunCarExample(*log.Logger, chan struct{})` runs whole solution creating multiple goroutines. 

In `./pkg/example/car_example.go` we have
```
fireupCarCommunication() - goroutine that communicates with car API library  
fireupIncreaseSpeed() - (task 3.) goroutine that consumes DataStore and send SetRequest (see type SetRequest struct in ./pkg/store/types.go) to fireupCarCommunication goroutine 
fireupPrintOutData() - (task 1.) goroutine that just prints out everything that it gets from carDataUpdateChan channel (see func RunCarExample in ./pkg/example/car_example.go)
fireupWriteDataIntoFile() - (task 2.) goroutine that creates file and writes to it all that it gets from carDataUpdateChan2 channel (see func RunCarExample in ./pkg/example/car_example.go) 
```

So here carCommunication is cleary owner of all DataStore object in this example. If we were to count all goroutines we get
```
fireupCarCommunication() +1 goroutine
fireupIncreaseSpeed() +1 goroutine
fireupPrintOutData() +1 goroutine
fireupWriteDataIntoFile() +1 goroutine
RunCarExample() +2 goroutines (WHY? because we create 2 goroutines inside this fn, these fn are Fanout) (see func RunCarExample in ./pkg/example/car_example.go)
```

And about task 4. ability to kill goroutines at any point of time. We have `var done chan struct{}` that I pass to every goroutine I create.   

# log15

This Package is a fork from [log15](https://github.com/inconshreveable/log15)

For more information, you can access the Source Link

## Add new features:
- Filter log level when writing into Log File .
- Rotating logfile automaticly when logfile size larger than setting or default. (eg: > 100M)
- You can rotate logfile whenever you want by using func "Rotate()"


### Usage:

## Importing

```go
import log "github.com/xiaomi-tc/log15"
```

## Examples

#### 1:  Basic usage

```go
func main() {

    h,_ := log.FileHandler("./app.log", log.LogfmtFormat())
    log.Root().SetHandler(h)

    Path := "http://woda.com"
    for i:=1; i < 1000000 ; i++ {
        log.Info("page accessed", "path", Path, "user_id", i)
        time.Sleep(50 * time.Millisecond)
    }
}
```

Will result in output that looks like this:

```
t=2017-09-26T01:40:37-0400 lvl=info msg="page accessed" path=http://woda.com user_id=21
t=2017-09-26T01:40:37-0400 lvl=info msg="page accessed" path=http://woda.com user_id=22
```


#### 2: Set LogLevel filter
```go
func main() {
    // set level to Warn
    log.SetOutLevel(log.LvlWarn)

    h,_ := log.FileHandler("./app.log", log.LogfmtFormat())
    log.Root().SetHandler(h)

    Path := "http://woda.com"
    for i:=1; i < 1000000 ; i++ {

        // Info is large than Warn, output none
        log.Info("page accessed", "path", Path, "user_id", i)
        time.Sleep(50 * time.Millisecond)
    }
}
```

Will output nothing.



###### Log level define: (from small to large)
- LvlCrit
- LvlError
- LvlWarn
- LvlInfo
- LvlDebug

#### 3: Modify default rotate parameters
```go
func main() {
    // set rotate parameters:
    // size:1m, keep 10days, backup 5 files, uncompress when rotate
    log.SetRotatePara(1,10,5,false)

    h,_ := log.FileHandler("./app.log", log.LogfmtFormat())
    log.Root().SetHandler(h)

    Path := "http://woda.com"
    for i:=1; i < 1000000 ; i++ {
        log.Info("page accessed", "path", Path, "user_id", i)
    }
```

Will result in output that looks like this:

```
log15]$ ls -l
total 8664
-rw-r--r--. 1 work work 1048524 Sep 26 02:43 app-2017-09-26T02-43-12.826.log
-rw-r--r--. 1 work work 1048524 Sep 26 02:43 app-2017-09-26T02-43-12.890.log
-rw-r--r--. 1 work work 1048524 Sep 26 02:43 app-2017-09-26T02-43-12.941.log
-rw-r--r--. 1 work work 1048524 Sep 26 02:43 app-2017-09-26T02-43-13.006.log
-rw-r--r--. 1 work work 1048524 Sep 26 02:43 app-2017-09-26T02-43-13.068.log
-rw-r--r--. 1 work work  667368 Sep 26 02:43 app.log
```

func "SetRotatePara()" define:
```
SetRotatePara(maxsize, maxage, maxbackup int, compress bool)
```
The default rotate  parameters:
- maxsize : 100  // 100M
- maxage: 10     // days
- maxbackup: 30  // files
- compress: true

#### 4: Force logfile rotate
```go
func main() {
    h,_ := log.FileHandler("./app.log", log.LogfmtFormat())
    log.Root().SetHandler(h)

    go func(){
      for {
          time.Sleep( 1 * time.Second)
          log.LogRotate()
      }
    }()

    Path := "http://woda.com"
    for i:=1; i < 1000000 ; i++ {
        log.Info("page accessed", "path", Path, "user_id", i)
        time.Sleep(50 * time.Millisecond)
    }
}
```

Will result in output that looks like this:

```
log15]$ ls -l
total 2924
-rw-r--r--. 1 work work     201 Sep 26 03:08 app-2017-09-26T03-08-59.134.log.gz
-rw-r--r--. 1 work work     171 Sep 26 03:09 app-2017-09-26T03-09-00.135.log.gz
-rw-r--r--. 1 work work     174 Sep 26 03:09 app-2017-09-26T03-09-01.136.log.gz
-rw-r--r--. 1 work work     171 Sep 26 03:09 app-2017-09-26T03-09-02.138.log.gz
-rw-r--r--. 1 work work     616 Sep 26 03:09 app.log
```
every second rotate once.

## License
Apache

### Integration tests for Supraworker
We create an API & supraworker consumes it

### Profile

```shell
go tool pprof -http=:8090  http://localhost:8088/debug/pprof/goroutine
```
### Check jobs data
```shell
docker exec -ti tests_db_1 mysql -uroot -ptest -D dev -e 'select * from jobs' 
```
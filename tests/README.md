### Integration tests for Supraworker
We create an API & supraworker consumes it

### Check jobs data
```shell
docker exec -ti tests_db_1 mysql -uroot -ptest -D dev -e 'select * from jobs' 
```
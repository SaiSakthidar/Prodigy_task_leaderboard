# Prodigy_task_leaderboard


## SETUP

```
docker-compose build
docker-compose up -d

```
## Login

```
http://localhost:8080/
```
## Create
```
curl -X POST http://localhost:8080/leaderboard \
  -H "Content-Type: application/json" \
  -H "X-API-Key: 8348hugirghhi9999jdja09kmvir" \
  -d '{
        "name": "test_user_1",
        "score": 400
      }'
```
## Update

```
curl -X PUT http://localhost:8080/leaderboard/ \
  -H "Content-Type: application/json" \
  -H "X-API-Key: 8348hugirghhi9999jdja09kmvir" \
  -d '{
        "name": "test_user_1",
        "score": 500
      }'
```

## Delete
```
curl -X DELETE http://localhost:8080/leaderboard/1 \
  -H "X-API-Key: 8348hugirghhi9999jdja09kmvir"
```

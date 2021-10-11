# MongoDB Docker

## Enter into container

```cmd
docker exec -it mongodb /bin/bash
```

## MongoDB Login

```cmd
mongo -u admin -p admin --authenticationDatabase admin
```


## Create User

```
db.createUser({user: 'ungame', pwd: 'secret', roles:[{'role': 'readWrite', 'db': 'workflows'}]});
```

## Create Database

```cmd
use workflows
```
## Login

```cmd
mongo -u ungame -p secret --authenticationDatabase workflows

use workflows

show collections
```
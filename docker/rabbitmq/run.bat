docker run -d --name rabbitmq ^
 -p 5672:5672 ^
 -p 15672:15672 ^
 --restart=always ^
 --hostname rabbitmq-master ^
 -v rabbitmq_volume:/var/lib/rabbitmq ^
 rabbitmq:3-management
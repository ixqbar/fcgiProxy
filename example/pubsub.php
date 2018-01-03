<?php

ini_set('default_socket_timeout', -1);

$redis_handle = new Redis();
$redis_handle->connect('127.0.0.1', 6899, 30);

$redis_handle->subscribe(array('*'), function($redis_handle, $chan, $message){
    echo $chan . '-' . $message . PHP_EOL;
});
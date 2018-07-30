<?php

$redis_handle = new Redis();
$redis_handle->connect('127.0.0.1', 6899);
$redis_handle->rawCommand("tpush", "*", '{"title":"pay","message":"nice"}');

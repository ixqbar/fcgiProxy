<?php

include '../src/push.lib.php';

Push::init('127.0.0.1', 6899);

echo (5 / 0) . PHP_EOL;

echo count($undefined_array) . PHP_EOL;

function test_func(int $num)
{
	return $num;
}

$list = [];

var_dump($list['name']);

print_r($todo);

echo test_func('haha') . PHP_EOL;
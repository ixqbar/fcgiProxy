<?php

$data = [
    'get'    => $_GET,
    'server' => $_SERVER,
    'raw'    => file_get_contents('php://input'),
];

//借助redis服务推送消息给指定客户（uuid标识）
try {
    $notification_handle = new Redis();
    $notification_handle->pconnect(
        '127.0.0.1',
        6899,
        30,
        'push'
    );
    $notification_handle->rawCommand('set', $_GET['uuid'], 'message from php');
} catch (Throwable $e) {

}

//返回接收到的数据
echo json_encode($data, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
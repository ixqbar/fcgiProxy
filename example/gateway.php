<?php

$data = [
    'get'    => $_GET,
    'server' => $_SERVER,
    'raw'    => file_get_contents('php://input'),
];
echo json_encode($data, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
<?php

print_r($_SERVER);
echo json_encode($_GET) . PHP_EOL;
echo file_get_contents('php://input') . PHP_EOL;
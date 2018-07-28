<?php

function curl($url, $post_data = '', $max_loop=1, $ext_header = array()) {
    $ch = curl_init();

    $options = array(
        CURLOPT_USERAGENT      => "GameCurl",
        CURLOPT_TIMEOUT        => 30,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_URL            => $url,
        CURLOPT_FOLLOWLOCATION => true,
        CURLOPT_PROXY          => '',
        CURLOPT_HTTPHEADER     => $ext_header ? array_merge(array("Expect:"), $ext_header) : array("Expect:"),
    );

    //post
    if ($post_data) {
        $options[CURLOPT_POST]       = true;
        $options[CURLOPT_POSTFIELDS] = $post_data;
    }

    static $curl_version = null;
    if (is_null($curl_version)) {
        $curl_version = curl_version();
    }

    curl_setopt_array($ch, $options);
    $result = curl_exec($ch);
    if (false === $result || curl_errno($ch)) {
        $max_loop++;
        if ($max_loop <= 3) {
            $result = curl($url, $post_data, $max_loop, $ext_header);
        }
    }
    curl_close($ch);

    return $result;
}

$content = [
    'noticeData' => [
        'category' => 'notice',
        'title'    => 'notice title',
        'message'  => 'notice message',
        'data'     => '',
    ],
    'noticeTime' => time(),
];

echo json_encode($content, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) . PHP_EOL;

$result = curl(
    'http://127.0.0.1:8899/push',
    json_encode($content, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES)
);

echo $result . PHP_EOL;
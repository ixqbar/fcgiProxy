<?php


function rc4($data, $pwd = 'hello') {
    $key[] ="";
    $box[] ="";
    $cipher = "";

    $pwd_length = strlen($pwd);
    $data_length = strlen($data);

    for ($i = 0; $i < 256; $i++) {
        $key[$i] = ord($pwd[$i % $pwd_length]);
        $box[$i] = $i;
    }

    for ($j = $i = 0; $i < 256; $i++) {
        $j = ($j + $box[$i] + $key[$i]) % 256;
        $tmp = $box[$i];
        $box[$i] = $box[$j];
        $box[$j] = $tmp;
    }

    for ($a = $j = $i = 0; $i < $data_length; $i++) {
        $a = ($a + 1) % 256;
        $j = ($j + $box[$a]) % 256;

        $tmp = $box[$a];
        $box[$a] = $box[$j];
        $box[$j] = $tmp;

        $k = $box[(($box[$a] + $box[$j]) % 256)];
        $cipher .= chr(ord($data[$i]) ^ $k);
    }

    return $cipher;
}

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
    'id'   => 123,
    'res'  => 'http://127.0.0.1/cdn/game.png',
    'type' => "res",
    'data' => 'not found'
];

$result = curl(
    'https://acc.us.mnf.rumblade.net:8899/logs?channel=haha',
    rc4(json_encode($content, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES)),
    1,
    ["Content-Type: application/octet-stream"]
);

echo $result . PHP_EOL;
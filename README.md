
### version 0.0.7

### description
```
客户端通过websocket可直接请求php-fpm下php代码，php可通过内置redis协议服务与客户端完成通信
```

### usage
```
fcgiProxy --config=config.xml
```
* redis      0.0.0.0:6899
* websokcet  ws://127.0.0.1:8899/sock[?uuid=指定连接客户端标记，可选&channel=推送频道，可选&proxy=是否代理默认否，可选]

### config
```
<?xml version="1.0" encoding="UTF-8" ?>
<config>
	<admin_server>0.0.0.0:6899</admin_server>
	<!-- http[s] -->
	<http_server>0.0.0.0:8899</http_server>
	<!-- ssl cert -->
	<http_ssl_cert>/data/ssl/server.crt</http_ssl_cert>
	<http_ssl_key>/data/ssl/server.key</http_ssl_key>
	<!-- enable http static file path -->
	<http_static_root>/data/resource</http_static_root>
	<!-- fastcgi -->
	<fcgi_server>0.0.0.0:9000</fcgi_server>
	<script_filename>/Users/xingqiba/workspace/php/zwj2-beta/zwj2/app/src/game_server/test/proxy.php</script_filename>
	<query_string><![CDATA[name=xingqiba&version=0.0.7]]></query_string>
	<header_params>
		<param>
			<key>ProxyVersion</key>
			<value>0.0.7</value>
		</param>
	</header_params>
	<!-- allow websocket origins to access -->
	<origins>*</origins>
	<!-- logs -->
	<logger>
		<!-- empty is no encrypt -->
		<rc4_encrypt_key>hello</rc4_encrypt_key>
		<!-- record logs to mysql-->
		<mysql>
			<ip>127.0.0.1</ip>
			<username>root</username>
			<password></password>
			<port>3306</port>
			<database>logs</database>
		</mysql>
	</logger>
</config>
```

### fcgi通过redis协议推送消息给客户端
```
set {uuid} message   #给指定客户端推送消息  uuid可通过$_SERVER['PROXY_UUID'] 获取 或者连接ws参数uuid指定
set * message        #给所有客户端推送消息
number               #获取在线人数
del *                #剔除所有在线客户端
del {uuid}           #剔除指定在线客户端， uuid可通过$_SERVER['PROXY_UUID'] 获取 或者连接ws参数uuid指定
subscribe channel       #订阅频道  系统默认频道*
publish channel message #发布消息到指定频道
```

### pub&sub
```
<?php

ini_set('default_socket_timeout', -1);

$redis_handle = new Redis();
$redis_handle->connect('127.0.0.1', 6899, 30);
$redis_handle->subscribe(array('*'), function($redis_handle, $chan, $message){
	//服务器间隔5秒发送PING
	if ($message == 'PING') return; 
	echo $chan . '-' . $message . PHP_EOL;
});
```

### http
```
curl 'http://127.0.0.1:8899/?format=json'  #获取统计信息
curl --data '{"id":1,"res":"hh","code":404,"info":"not found"}' 'http://127.0.0.1:8899/logs' #提交日志
```
* http://xxx.xxx.xxx.xxx:pppp/
* http://xxx.xxx.xxx.xxx:pppp/logs[?channel=推送频道，可选]
* http://xxx.xxx.xxx.xxx:pppp/res

更多疑问请+qq群 233415606

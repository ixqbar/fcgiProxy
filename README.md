
### version 0.0.1

### description
```
客户端通过websocket可直接请求php-fpm下php代码，php可通过内置redis协议服务与客户端完成通信
```

### usage
```
fcgiProxy --config=config.xml
```
* redisproxy 0.0.0.0:6899
* websokcet  ws://127.0.0.1:8899/proxy

### config
```
<?xml version="1.0" encoding="UTF-8" ?>
<config>
	<admin_server>0.0.0.0:6899</admin_server>
	<http_server>0.0.0.0:8899</http_server>
	<php_server>0.0.0.0:9000</php_server>
	<script_filename>/Users/xingqiba/workspace/php/zwj2-beta/zwj2/app/src/game_server/test/proxy.php</script_filename>
	<query_string><![CDATA[name=xingqiba&version=0.0.1]]></query_string>
	<header_params>
		<param>
			<key>PHP-Proxy</key>
			<value>1</value>
		</param>
	</header_params>
</config>
```

### php通过redis协议推送消息给客户端
```
set {uuid} message   #给指定客户端推送消息  uuid可通过$_SERVER['PROXY_UUID'] 获取
set * message        #给所有客户端推送消息
```

更多疑问请+qq群 233415606

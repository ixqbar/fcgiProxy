<?php

class Push
{
	static $host = '127.0.0.1';
	static $port = 6899;

	/**
	 * @param string $host
	 * @param int $port
	 * @param bool $enable_push_errors
	 */
	public static function init(string $host, int $port, bool $enable_push_errors = true)
	{
		self::$host = $host;
		self::$port = $port;

		if ($enable_push_errors) {
			set_error_handler(function (int $err_no, string $err_msg, string $err_file = '', int $err_line = 0, array $err_context = []) {
				self::send(sprintf('%s in file %s at line %s', $err_msg, $err_file, $err_line), 'ERROR');
			});

			set_exception_handler(function(Throwable $e){
				self::send($e->getMessage() . ',FILE:' . $e->getFile() . ',LINE:' . $e->getLine(), 'ERROR');
			});

			register_shutdown_function(function () {
				try {
					$last_error = error_get_last();
					if (empty($last_error)) {
						return;
					}

					$message = '[' . $last_error['type'] . ']' . $last_error['message'] . ' in ' . $last_error['file'] . ' at ' . $last_error['line'] . ' line';
					self::send($message, 'ERROR');
				} catch (Throwable $e) {
					error_log($e->getTraceAsString());
				}
			});

		}
	}

	static $push_handle = null;

	/**
	 * @param string $message
	 * @param string|null $title
	 */
	public static function send(string $message, string $title = null)
	{
		try {
			if (self::$push_handle == null) {
				self::$push_handle = new Redis();
				self::$push_handle->pconnect(self::$host, self::$port, 10, 'push');
			}
			$notification = [
				'title'   => $title ?? date('Y-m-d H:i:s'),
				'message' => $message,
			];
			self::$push_handle->rawCommand('tpush', '*', json_encode($notification, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES));
		} catch (Throwable $e) {
			self::$push_handle = null;
		} finally {

		}
	}

}
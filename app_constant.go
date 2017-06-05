package main

const(
	ONLINE_PAGE_SERVER_LIST = "https://www.online.net/fr/serveur-dedie"
	MESSAGE_SERVER_AVAILABILITY = "NEW [%s] server - %d available(s) (previous %d)"
	MESSAGE_SERVER_AVAILABILITY_REDUCE = "Server [%s] availability reduced - %d available(s) (previous %d)"
	MESSAGE_SERVER_STATUS = "[%s] %d available(s)"
	DEFAULT_CHANNEL_NAME = "server-availability-notify"

	COMMAND_LIST_SERVER = "list"
	COMMAND_SUBSCRIBE_SERVER_UPDATE = "subscribe"

	MESSAGE_SUBSCRIBE_SERVER = "Registered to [%s] updates"
	MESSAGE_SERVER_NOT_FOUND = "Server name [%s] not found"
	MESSAGE_NOTIFY_SERVER_UPDATE = "@%s [%s] avaibility update, %d available(s) (previous %d)"
)

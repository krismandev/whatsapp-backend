# Whatsapp Backend Service (wabackend)

wabackend is a service that help manage the data flow of wapp-nodejs by passing requests from waapihandler to wapp-nodejs, and request from wapp-nodejs to wamessagehandler. This service responsible with all wapp-nodejs pods. This service contain many background process that always listen to new event from wapp-nodejs

### Installation
First you need to setup the configuration in ./config/config.yml. Make sure all the database map in dblist is filled correctly. 
```bash
#if you want to run in local environtment
go mod vendor
go run main.go
```
```bash
#if you want to build the image
sh builddocker.sh {version}
#example : 
sh builddocker.sh v1.0.0
```
We will use the v[MAJOR].[MINOR].[PATCH] format for every pushed update in git. After image builded, you can update the image version used in kubernetes yaml in these paths : 
- [DEV] /mnt/gfs/devdist/yaml/whatsapp/wabackend.yml
- [PROD] /mnt/gfs/config/prod/app/whatsapp/wabackend.yml


### Code Folder Structure
- config : Contains yaml configuration that will be used in apps
- commonlib & lib : Contains third party / helper functions to help develop the apps
- connections : Contains configuration to connect to MySQL, Redis or another databases
- datastruct : Contains all module request & response structure.
- model : Contains function with direct database logic. (Example : get data, insert/update data, delete data)
- route : Contains HTTP endpoint routing to transport 
- services : Contains request logic for every module
- transport : Contains request encoder, decoder, and endpoint definition

### Available Services 

##### Send Direct Chat
(Request_service.go **DirectChat()**)
By default all chat flow is use an async flow. But this endpoint is used by waapihandler to create sync request to wapp-nodejs, and return the response to waapihandler.

##### Record The NodeJS Pod Active Status
(Request_service.go **NodeNotify()**)
wapp-nodejs will always send a live signal to wabackend to mark the pod as active. This endpoint will record the pod notify status to redis.

##### Get Available wapp-nodejs Node Metrics
(Request_service.go **NodeMetrics()**)
This endpoint will return all available wapp-nodejs pod. This service get the available pod data from redis that is stored by NodeNotify() endpoint. This live metric is will be shown to our internal php-mdashboard in menu "Chatbots > Whatsapp Pod Status"

##### Call Wapp-NodeJS 
(Request_service.go **CallFrontend()**)
This endpoint is called from waapihandler to trigger calling the wapp-nodejs controller. We can call connect, disconnect, generate-history-chat, retract-message.

##### Handle Bot State Changes
(Request_service.go **DeviceStatusRequest()**)
This endpoint will be called from wapp-nodejs with device state information, and this endpoint will store the latest changes to redis and notify to waeventlog service. When bot device is disconnected or timeout, this endpoint is the first place that get the request to be passed to another service.

##### Store Media Image/Document Data
(Request_service.go **PassMedia()**)
This endpoint will be called from wapp-nodejs when there is incoming/outgoing message with media data such as image/document/video to. All of media data will be passed to this endpoint to be stored to local storage. So when we open php-whatsapp chat dialog, we can open and download the media.

##### Get Contacts
(Request_service.go **GetContacts()**)
This endpoint will be called from waapihandler to get the latest contact list in targetted bot

##### Chat History Generation
(History_service.go **CallHistoryGeneration() StoreHistoryGeneration()**)
CallHistoryGeneration will be triggered from wapp-nodejs after successfully connect. After that, wapp-nodejs will call StoreHistoryGeneration to store the history chat data. This history chat data is used in php-whatsapp only, so when customer open the application, the old chat data will be shown.


### Background Processes

##### processors.BotTypeMapping()
This background process only run on first runtime & occassionally only. This process will help map the bot number to wapp-nodejs pod types. Until now, we have a several wapp-nodejs pod instance with different name such as : API, CS, OTP, APP. This process will create a map to relate the bot number to which bot type.

##### processors.ManageNodesHealth()
This background process will running infinitely and sleep every 1 minute. This process will check if there is an unactive wapp-nodejs pod, then that pod will be removed from redis, and bot status in that pods will be updated to "Disconnect". 

##### processors.BotAutoReconnect()
This background process only run on first runtime & occassionally only. When triggered, this process will read the bot to pod map data, and will call wapp-nodejs reconnect to make sure the bot will be auto reconnect. With this background process, when we restart the wabackend pod all nodejs instance will be automatically reconnected because this process is running on first runtime too. 

##### processors.QueueListener()
This background process running infinitely listen for outgoing message request queue from waapihandler. This process will send message request to wapp-nodejs, and store the wapp-nodejs response to MessageData & MessageState queue. (Will be processed by wamsgprocessor service, so the delivery status can be grabbed from wadrprocesor service)

##### processors.IncomingMessageListener()
This background process running infinitely listen to message queue from wapp-nodejs. When there are new incoming message event from wapp-nodejs, all chat data will be pushed to queue and processed here. After processed, the chat data will be passed to MessageData & MessageState too. wamsgprocessor will handle all of them, so if the bot have autoreply webhook endpoint can be processed by wabothandler.

##### processors.IncomingEventListener()
This background process running infinitely listen to event queue from wapp-nodejs. If there is an event such as QRReceived, Authenticated, AuthFailed, then the event will be passed to waapihandler. If there is an event SendMsg detected, then this process will update the ACK (Acknowledgement) of the message to websocket/wadata service. 


### FAQ

package processors

import (
	"encoding/json"
	"errors"
	"skeleton/config"
	"skeleton/connections"
	"skeleton/datastruct"
	"skeleton/lib"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func ManageNodesHealth(conn *connections.Connections, wg *sync.WaitGroup, closeFlag <-chan struct{}) {
	wg.Add(1)

	defer wg.Done()
	tick := time.NewTicker(time.Duration(1) * time.Minute)
	for {
		select {
		case <-closeFlag:
			return
		case <-tick.C:
			logrus.Infof("ManageNodesHealth CALLED")
			nodeList, _ := conn.DBRedis.Get("NODE-POD-LISTS").Result()
			if len(nodeList) == 0 {
				logrus.Infof("ManageNodesHealth : No whatsapp node pods available right now")
				time.Sleep(time.Duration(1) * time.Minute)
				continue
			}
			splitNode := strings.Split(nodeList, "|")
			changes := false
			for _, hostname := range splitNode {
				timestamp, _ := conn.DBRedis.Get("NODE-POD-" + hostname + "-TIMESTAMP").Result()
				timeinstance, _ := time.Parse("2006-01-02 15:04:05", timestamp)
				skrg := time.Now().Format("2006-01-02 15:04:05")
				skrginstance, _ := time.Parse("2006-01-02 15:04:05", skrg)
				timeunix := skrginstance.Unix() - timeinstance.Unix()
				if int(timeunix) > config.Param.MaximumTimeoutThreshold {
					logrus.Errorf("NODE POD %s STATUS NOT UPDATED FOR MORE THAN %d SECONDS. POD WILL BE SHUTTED DOWN", hostname, config.Param.MaximumTimeoutThreshold)
					conn.DBRedis.Set("NODE-POD-"+hostname+"-BOT-CONNECTED", "0", 0)
					conn.DBRedis.Del("NODE-POD-" + hostname + "-BOTS")
					changes = true
				}
			}

			if changes {
				logrus.Infof("Node changes detected! Run ManageDisconnectedBotInEveryNode to remove inactive bots in pod")
				ManageDisconnectedBotInEveryNode(conn, "0")
			}

			// logic ini akan diulang2 terus utk memastikan node yg tidak aktif langsung dihapus
			// time.Sleep(time.Duration(1) * time.Minute)
		}
	}
}

func NodeUpdateConfig(conn *connections.Connections, req datastruct.NodeNotifyRequest) {
	var err error
	var activePods []string

	activePodStrings, _ := conn.DBRedis.Get("NODE-POD-LISTS").Result()
	if len(activePodStrings) > 0 {
		activePods = strings.Split(activePodStrings, "|")
	}

	_, inSlice := lib.FindSlice(activePods, req.Host)
	if !inSlice {
		// pod belum terdaftar, update key pod-lists register
		activePods = append(activePods, req.Host)
		conn.DBRedis.Set("NODE-POD-LISTS", strings.Join(activePods, "|"), 0)
	}

	// update last pod timestamp & bot list
	botType := req.BotType
	if len(botType) == 0 {
		botType = "cs"
	}

	currentTimestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = conn.DBRedis.Set("NODE-POD-"+req.Host+"-STATE", "CONNECT", 0).Result()
	_, err = conn.DBRedis.Set("NODE-POD-"+req.Host+"-TIMESTAMP", currentTimestamp, 0).Result()
	_, err = conn.DBRedis.Set("NODE-POD-"+req.Host+"-TYPE", botType, 0).Result()
	_, err = conn.DBRedis.Set("NODE-POD-"+req.Host+"-PORT", req.Port, 0).Result()
	_, err = conn.DBRedis.Set("NODE-POD-"+req.Host+"-BOT-CONNECTED", strconv.Itoa(len(req.CurrentActiveBot)), 0).Result()
	_, err = conn.DBRedis.Set("NODE-POD-"+req.Host+"-BOTS", strings.Join(req.CurrentActiveBot, "|"), 0).Result()
	if err != nil {
		logrus.Errorf("Error Redis SET in NodeUpdateConfig : %+v", err)
	}

	nodeFirstTime := req.FirstTime
	ManageDisconnectedBotInEveryNode(conn, nodeFirstTime)
}

// ManageDisconnectedBotInEveryNode dipanggil saat ada penambahan / pengurangan bot, penambahan / pengurangan node
func ManageDisconnectedBotInEveryNode(conn *connections.Connections, firstTime string) {
	var storedConnectedBots []string
	var currentConnectedBots []string

	activeBotLists, _ := conn.DBRedis.Get(config.Param.RedisPrefix + "ListBotStatus").Result()
	if len(activeBotLists) > 0 {
		storedConnectedBots = strings.Split(activeBotLists, "|")
	}

	var needRefreshConfig = false

	var splitNode []string
	nodeList, _ := conn.DBRedis.Get("NODE-POD-LISTS").Result()
	if len(nodeList) > 0 {
		splitNode = strings.Split(nodeList, "|")
	}

	for _, nname := range splitNode {
		bots, _ := conn.DBRedis.Get("NODE-POD-" + nname + "-BOTS").Result()
		if len(bots) > 0 {
			splitBot := strings.Split(bots, "|")
			for _, bot := range splitBot {
				// assign current bot active ke variabel ini supaya nanti bisa dicari tahu jika ada yg disconnects
				currentConnectedBots = append(currentConnectedBots, bot)

				// jika bot belum ada di storedConnectedBots : assign ke storedConnectedBots & set redis key connect=1
				if _, inSlice := lib.FindSlice(storedConnectedBots, bot); !inSlice {
					logrus.Infof("TRIGGER UPDATE MSGCONFIG BotStatus-%s : Connected", bot)
					StoreMsgConfig(conn, config.Param.RedisPrefix+"BotStatus-"+bot, "1")
					needRefreshConfig = true
				}
			}
		}
	}

	for _, storedBot := range storedConnectedBots {
		if _, inSlice := lib.FindSlice(currentConnectedBots, storedBot); !inSlice {
			// jika storedBot tidak ada di list currentConnectedBots, artinya nomor tsb sudah terdiskonek entah krn sesuatu
			logrus.Infof("TRIGGER UPDATE MSGCONFIG BotStatus-%s : Disconnect", storedBot)
			StoreMsgConfig(conn, config.Param.RedisPrefix+"BotStatus-"+storedBot, "0")
			needRefreshConfig = true

			// jika firstTime == 1, langsung coba retry reconnect
			// CARA RECONNECT = set redis counter reconnect key ke 0 lagi
			conn.DBRedis.Set("BOT-RECONNECT-COUNTER-"+storedBot, "0", 0)
		}
	}

	// update ListBotStatus
	conn.DBRedis.Set(config.Param.RedisPrefix+"ListBotStatus", strings.Join(currentConnectedBots, "|"), 0)
	logrus.Infof("Set ListBotStatus = %s", strings.Join(currentConnectedBots, "|"))

	if needRefreshConfig {
		// trigger refresh config
		ReloadMsgConfig()

		// jalankan auto reconnect jika firstTime == 1
		if firstTime == "1" {
			BotAutoReconnect(conn, false, &lib.WaitG)
		}
	} else {
		// logrus.Infof("ManageDisconnectedBotInEveryNode CALLED BUT NOTHING TO DO")
	}

}

// ini adalah fungsi yg harus selalu dipanggil setiap kali update config atau setiap kali ada perubahan event
func BotTypeMapping(conn *connections.Connections, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	stringUserList, _ := GetMsgConfig(conn, config.Param.RedisPrefix+"UserLists")
	if len(stringUserList) == 0 {
		logrus.Errorf("ERROR NodeMapping CANNOT BE DONE. NO USER LISTS DETECTED FROM MSGCONFIG : WAUserLists")
		return
	}

	users := strings.Split(stringUserList, "|")
	var botMap = make(map[string]string)

	for _, userID := range users {
		stringBotList, _ := GetMsgConfig(conn, config.Param.RedisPrefix+"UserBotList-"+userID)
		if len(stringBotList) == 0 {
			// skipped, no bot in this username
			continue
		}

		splitBot := strings.Split(stringBotList, "|")
		for _, botNumber := range splitBot {
			botapp, _ := GetMsgConfig(conn, config.Param.RedisPrefix+"AppID-"+botNumber)
			if botapp == "5" { //appid whatsapp
				botType, _ := GetMsgConfig(conn, config.Param.RedisPrefix+"BotType-"+botNumber)
				if _, inslice := lib.FindSlice([]string{"cs", "api", "otp"}, strings.ToLower(botType)); !inslice {
					botType = "cs" // fallback
				}
				botMap[botNumber] = botType
			} else {
				// skip all bot that is not whatsapp
				continue
			}
		}
	}

	botMarshal, err := json.Marshal(botMap)
	if err != nil {
		logrus.Infof("Failed to marshal the botMap in NodeMapping : %+v", err)
	}
	_, err = conn.DBRedis.Set("BOT-TYPE-MAP", string(botMarshal), 0).Result()
	if err != nil {
		logrus.Errorf("Cannot store BOT-TYPE-MAP in NodeMapping : %+v", err)
	}
}

func NodeSelector(conn *connections.Connections, botNumber string, selectNew bool) (result string, err error) {
	logrus.Infof("Call NodeSelector to get available node for bot number %s", botNumber)
	if !selectNew {
		// ambil dari redis yg digenerate pertama kali saat konek
		result, err = conn.DBRedis.Get("WANODE-SCHEDULED-" + botNumber).Result()
		if len(result) > 0 {
			logrus.Infof("NodeSelector try get scheduled value from cache : %s", result)
			return
		}
	} else {
		// hapus key cache node scheduled
		conn.DBRedis.Del("WANODE-SCHEDULED-" + botNumber)
	}

	nodeStringMap, err := conn.DBRedis.Get("BOT-TYPE-MAP").Result()
	if err != nil {
		logrus.Errorf("Error Redis in NodeSelector : %+v", err)
		return
	}

	botMap := make(map[string]string)
	err = json.Unmarshal([]byte(nodeStringMap), &botMap)
	if err != nil {
		logrus.Errorf("Failed to unmarshal botMap in NodeSelector : %+v", err)
		return
	}

	usedBotType, found := botMap[botNumber]
	logrus.Infof("BOTMAP : %+v", botMap)
	if !found {
		// nomor ini nggak dikenal, jadi nggak tahu bakal dimasukin ke channel apa. default : CS
		err = errors.New("botNumber " + botNumber + " is unknown. cannot schedule the node pod")
		logrus.Errorf("%+v", err)
		return
	}
	result, err = GetAvailableNodeByType(conn, usedBotType)
	if len(result) > 0 {
		conn.DBRedis.Set("WANODE-SCHEDULED-"+botNumber, result, 0)
	}

	return
}

func GetAvailableNodeByType(conn *connections.Connections, botType string) (result string, err error) {
	podList, err := conn.DBRedis.Get("NODE-POD-LISTS").Result()
	if len(podList) == 0 {
		err = errors.New("No node pod available right now")
		return
	}

	var matchedPodType = make(map[string]int)
	var lowestPodType = 0

	splitPod := strings.Split(podList, "|")
	maximumTimeoutThreshold := config.Param.MaximumTimeoutThreshold
	var updatedSplitPod []string

	for _, hostname := range splitPod {
		if len(hostname) == 0 {
			continue
		}
		timestamp, _ := conn.DBRedis.Get("NODE-POD-" + hostname + "-TIMESTAMP").Result()
		connectionCount, _ := conn.DBRedis.Get("NODE-POD-" + hostname + "-BOT-CONNECTED").Result()
		nodeType, _ := conn.DBRedis.Get("NODE-POD-" + hostname + "-TYPE").Result()
		if len(timestamp) > 0 {
			timeinstance, err := time.Parse("2006-01-02 15:04:05", timestamp)
			if err != nil {
				logrus.Errorf("Cannot parse the timestamp in NODE-POD-%s-TIMESTAMP key : %s", hostname, timestamp)
				continue
			}

			timeunix := time.Now().Unix() - timeinstance.Unix()
			if int(timeunix) > maximumTimeoutThreshold {
				logrus.Errorf("NODE POD %s STATUS NOT UPDATED FOR MORE THAN 5 MINUTES. POD WILL BE SHUTTED DOWN", hostname)
				conn.DBRedis.Del("NODE-POD-" + hostname + "-TIMESTAMP")
				conn.DBRedis.Del("NODE-POD-" + hostname + "-BOT-CONNECTED")
			} else {
				// pod masih dianggap aktif, masukkan ke map updatedSplitPod untuk disimpan ulang jika ada perubahan state node pod
				updatedSplitPod = append(updatedSplitPod, hostname)

				if strings.ToLower(nodeType) == botType {
					cc, _ := strconv.Atoi(connectionCount)
					if lowestPodType == 0 || cc < lowestPodType {
						lowestPodType = cc
					}
					matchedPodType[hostname] = cc
				}
			}
		} else {
			// skipped. broken redis key structure
		}
	}

	// save new updatedSplitPod if there is node changes (kemungkinan karena ada pod yg shutdown)
	if len(updatedSplitPod) != len(splitPod) {
		conn.DBRedis.Set("NODE-POD-LISTS", strings.Join(updatedSplitPod, "|"), 0)
		ManageDisconnectedBotInEveryNode(conn, "0")
	}

	// cek matchedPodType. cari yang angkanya paling kecil
	for hostname, ccount := range matchedPodType {
		if ccount <= lowestPodType {
			// use this hostname
			result = hostname
			return
		}
	}

	// kalau masih nggak ketemu : kasih notif bahwa tidak ada pod tersedia
	err = errors.New("No pods available for this botType")
	logrus.Infof("Error node pod %s not found : %+v", botType, err)
	return

}

// normalnya method ini ditrigger saat : init pertama kali, dan saat ada pod yg mati utk trigger auto scheduling
func BotAutoReconnect(conn *connections.Connections, firstInit bool, wg *sync.WaitGroup) {
	wg.Add(1)

	defer wg.Done()
	var botMapData = make(map[string]string)
	botMap, _ := conn.DBRedis.Get("BOT-TYPE-MAP").Result()
	err := json.Unmarshal([]byte(botMap), &botMapData)
	if err != nil {
		logrus.Errorf("Error unmarshall BOT-TYPE-MAP : %+v", err)
	}

	if firstInit {
		// reset seluruh bot number increment = 0
		for botNum := range botMapData {
			conn.DBRedis.Set("BOT-RECONNECT-COUNTER-"+botNum, "0", 0)
		}
	}

	autoconnectThreshold := 1

	for botNum, botType := range botMapData {
		reconnectCounter, _ := conn.DBRedis.Get("BOT-RECONNECT-COUNTER-" + botNum).Result()
		nCounter, _ := strconv.Atoi(reconnectCounter)
		if nCounter < autoconnectThreshold {
			// try reconnect this number
			nodeTarget, err := GetAvailableNodeByType(conn, botType)
			if err != nil {
				logrus.Errorf("Cannot get GetAvailableNodeByType in BotAutoReconnect for phone number %s : %+v", botNum, err)
				continue
			}

			// call frontend
			port, _ := conn.DBRedis.Get("NODE-POD-" + nodeTarget + "-PORT").Result()
			if len(port) == 0 {
				port = config.Param.WAClientPort
			}
			finalURL := "http://" + nodeTarget + config.Param.WAClientNamespace + port + "/app/connect?reconnect=OK"
			logrus.Infof("TRY AUTOCONNECT FRONTEND (BOT = %s) WITH FINAL URL : %s", botNum, finalURL)
			var pass = datastruct.RequestJSONRequest{
				Phone: botNum,
			}

			httpBody, httpCode := lib.CallRestAPIOnBehalf(pass, finalURL, "GET", "", 5*time.Second)
			if httpCode >= 200 && httpCode < 400 {
				// success called
				logrus.Infof("AUTOCONNECT FOR BOTNUM %s SUCCESSFULLY CALLED : %s", botNum, httpBody)
			}
			nCounter = nCounter + 1
			conn.DBRedis.Set("BOT-RECONNECT-COUNTER-"+botNum, strconv.Itoa(nCounter), 0)

		} else {
			// this botNumber is already reconnected. skip
		}
	}

	// reconnect
}

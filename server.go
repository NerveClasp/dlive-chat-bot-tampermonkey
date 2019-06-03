package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// IncomingJSON json
type IncomingJSON struct {
	Data Data `json:"data"`
}

// Data json
type Data struct {
	UserByDisplayName UserByDisplayName `json:"userByDisplayName"`
}

// UserByDisplayName json
type UserByDisplayName struct {
	Chats []Chat `json:"chats"`
}

// Chat incoming json struct
type Chat struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Sender  Sender `json:"sender"`
}

// Sender incoming json struct
type Sender struct {
	Displayname string `json:"displayname"`
	Avatar      string `json:"avatar"`
}

// ChatMessage chat struct
type ChatMessage struct {
	Sender  string `json:"sender"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
}

// ChatMessages array of ChatMessage
type ChatMessages struct {
	mu   sync.RWMutex
	last []ChatMessage
}

// Store stores new array of ChatMessage
func (d *ChatMessages) Store(data []ChatMessage) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.last = data
}

// Get returns current []ChatMessage
func (d *ChatMessages) Get() []ChatMessage {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.last
}

// Config the structure of config.sjon
type Config struct {
	mu            sync.RWMutex
	UserName      string `json:"userName"`
	LastNMessages int    `json:"lastNMessages"`
	Port          int    `json:"port"`
}

// GetUser returns user name from config
func (c *Config) GetUser() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.UserName
}

// GetLastNMessages returns the number of last messages to show
func (c *Config) GetLastNMessages() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.LastNMessages
}

// GetPort returns the port to run the server on
func (c *Config) GetPort() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Port
}

// GetPortString returns the port as string
func (c *Config) GetPortString() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return strconv.Itoa(c.Port)
}

func main() {
	fs := http.FileServer(http.Dir("chat"))
	http.Handle("/", fs)
	configRaw, err := ioutil.ReadFile("config.json")
	check(err)
	var config Config
	json.Unmarshal(configRaw, &config)
	port := config.GetPortString()
	var chats ChatMessages
	chatsHandler := makeChatsHandler(&chats)
	http.HandleFunc("/api/messages", chatsHandler)
	interval := time.NewTicker(3 * time.Second)
	go func() {
		for range interval.C {
			fetch(&config, &chats)
			// write(chats)
		}
	}()

	log.Println("Listening on http://localhost:" + port)
	http.ListenAndServe(":"+port, nil)
}

func makeChatsHandler(chatMessages *ChatMessages) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		profile := chatMessages.Get()
		js, err := json.Marshal(profile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// func write(chats ChatMessages) {
// 	// To start, here's how to dump a string (or just
// 	// bytes) into a file.
// 	f, err := os.OpenFile("chat.txt", os.O_RDWR|os.O_CREATE, 0755)
// 	check(err)
// 	chatString, _ := json.Marshal(chats)
// 	d1 := []byte(chatString)
// 	fmt.Println(string(chats[0].Avatar))
// 	if _, err := f.Write(d1); err != nil {
// 		log.Fatal(err)
// 	}
// 	if err := f.Close(); err != nil {
// 		log.Fatal(err)
// 	}
// }

func fetch(config *Config, lastSix *ChatMessages) {
	displayname := config.GetUser()
	lastN := config.GetLastNMessages()
	var jsonStr = []byte(`{"operationName":"LivestreamChatroomInfo","variables":{"displayname":"` + displayname + `","isLoggedIn":false,"limit":20},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"c38d67b66455636fee3e0c3f96e5aa53cf344cc99386def56de6b51a27ae36a7"}}}`)
	req, err := http.NewRequest("POST", "https://graphigo.prd.dlive.tv/", bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Fingerprint", "")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referrer", "https://dlive.tv/"+displayname)
	req.Header.Set("ReferrerPolicy", "no-referrer-when-downgrade")
	req.Header.Set("Mode", "cors")

	client := &http.Client{}
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var dat IncomingJSON
	json.Unmarshal(body, &dat)
	chats := dat.Data.UserByDisplayName.Chats
	messages := make([]ChatMessage, 0)
	lastMessages := make([]ChatMessage, 0)
	for j := 0; j < len(chats); j++ {
		Type := chats[j].Type
		if Type == "Message" {
			messages = append(messages, ChatMessage{
				Sender:  chats[j].Sender.Displayname,
				Avatar:  chats[j].Sender.Avatar,
				Message: chats[j].Content,
			})
		}
	}
	if len(messages) >= lastN {
		for k := len(messages) - lastN; k < len(messages); k++ {
			lastMessages = append(lastMessages, messages[k])
		}
		lastSix.Store(lastMessages)
	} else {
		lastSix.Store(messages)
	}
}

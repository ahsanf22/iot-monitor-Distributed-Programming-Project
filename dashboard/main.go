package main

import (
	"html/template"
	"log"
	"net/http"
	"sync"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type SensorData struct {
	Temp  string
	Humid string
	Run   bool
}

type SystemState struct {
	S1 SensorData
	S2 SensorData
}

var state = SystemState{
	S1: SensorData{Temp: "--", Humid: "--", Run: true},
	S2: SensorData{Temp: "--", Humid: "--", Run: true},
}
var mqttClient mqtt.Client
var mutex sync.Mutex

// --- HTML: LOGIN ---
const loginHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Login</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: sans-serif; text-align: center; padding: 20px; background: #eee; }
        .box { background: white; padding: 30px; border-radius: 10px; box-shadow: 0 4px 8px #ccc; max-width: 400px; margin: auto; }
        input { padding: 10px; margin: 10px 0; width: 90%; font-size: 16px; }
        button { padding: 10px 20px; background: #007bff; color: white; border: none; cursor: pointer; font-size: 16px; width: 100%; }
        .error { color: red; margin-bottom: 10px; }
    </style>
</head>
<body>
    <div class="box">
        <h2>IoT Secure Login</h2>
        {{if .}}<div class="error">{{.}}</div>{{end}}
        <form method="post" action="/login">
            <input type="text" name="username" placeholder="Username" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Login</button>
        </form>
        <p style="color:#888;">Try: <b>admin / admin123</b></p>
    </div>
</body>
</html>`

// --- HTML: DASHBOARD ---
const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Dashboard</title>
    <meta http-equiv="refresh" content="2">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: sans-serif; text-align: center; background: #222; color: white; padding: 20px; margin: 0; }
        .container { display: flex; justify-content: center; gap: 20px; flex-wrap: wrap; }
        
        .card { 
            border: 3px solid #555; 
            padding: 20px; 
            border-radius: 15px; 
            background: #333; 
            width: 100%; max-width: 350px; 
            transition: all 0.3s ease;
        }

        .data-row { display: flex; justify-content: space-around; margin: 15px 0; }
        .val { font-size: 30px; font-weight: bold; }
        .label { font-size: 12px; color: #aaa; text-transform: uppercase; }

        .status-running { border-color: #2ecc71; }
        .text-running { color: #2ecc71; }
        .status-stopped { border-color: #e74c3c; }
        .text-stopped { color: #e74c3c; }

        .btn { padding: 15px; border: none; border-radius: 5px; cursor: pointer; font-weight: bold; width: 100%; margin-top: 10px; font-size: 16px; }
        .btn-stop { background: #e74c3c; color: white; }
        .btn-start { background: #2ecc71; color: white; }
    </style>
</head>
<body>
    <h1>IoT Monitor</h1>
    <div class="container">
        
        <div class="card {{if .State.S1.Run}}status-running{{else}}status-stopped{{end}}">
            <h3>Living Room</h3>
            <div class="data-row {{if .State.S1.Run}}text-running{{else}}text-stopped{{end}}">
                <div>
                    <div class="val">{{.State.S1.Temp}}°C</div>
                    <div class="label">Temperature</div>
                </div>
                <div>
                    <div class="val">{{.State.S1.Humid}}%</div>
                    <div class="label">Humidity</div>
                </div>
            </div>
            {{if .IsAdmin}}
            <form action="/toggle" method="post">
                <input type="hidden" name="id" value="1">
                {{if .State.S1.Run}}
                    <button class="btn btn-stop">STOP SENSOR</button>
                {{else}}
                    <button class="btn btn-start">START SENSOR</button>
                {{end}}
            </form>
            {{end}}
        </div>

        <div class="card {{if .State.S2.Run}}status-running{{else}}status-stopped{{end}}">
            <h3>Bedroom</h3>
            <div class="data-row {{if .State.S2.Run}}text-running{{else}}text-stopped{{end}}">
                <div>
                    <div class="val">{{.State.S2.Temp}}°C</div>
                    <div class="label">Temperature</div>
                </div>
                <div>
                    <div class="val">{{.State.S2.Humid}}%</div>
                    <div class="label">Humidity</div>
                </div>
            </div>
            {{if .IsAdmin}}
            <form action="/toggle" method="post">
                <input type="hidden" name="id" value="2">
                {{if .State.S2.Run}}
                    <button class="btn btn-stop">STOP SENSOR</button>
                {{else}}
                    <button class="btn btn-start">START SENSOR</button>
                {{end}}
            </form>
            {{end}}
        </div>

    </div>
    <br><br>
    <a href="/logout" style="color:#aaa;">Log Out</a>
</body>
</html>`

// --- GO LOGIC ---
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		u := r.FormValue("username")
		p := r.FormValue("password")
		if u == "admin" && p == "admin123" {
			http.SetCookie(w, &http.Cookie{Name: "role", Value: "admin", Path: "/"})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else if u == "user" && p == "user123" {
			http.SetCookie(w, &http.Cookie{Name: "role", Value: "user", Path: "/"})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		t, _ := template.New("login").Parse(loginHTML)
		t.Execute(w, "❌ Incorrect Username or Password")
		return
	}
	t, _ := template.New("login").Parse(loginHTML)
	t.Execute(w, nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("role")
	if err != nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	mutex.Lock()
	d := state
	mutex.Unlock()
	t, _ := template.New("dash").Parse(dashboardHTML)
	t.Execute(w, struct{ State SystemState; IsAdmin bool }{d, c.Value == "admin"})
}

func toggleHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	mutex.Lock()
	if id == "1" {
		cmd := "START"; if state.S1.Run { cmd = "STOP" }
		mqttClient.Publish("home/sensor1/cmd", 0, false, cmd)
		state.S1.Run = !state.S1.Run
	} else {
		cmd := "START"; if state.S2.Run { cmd = "STOP" }
		mqttClient.Publish("home/sensor2/cmd", 0, false, cmd)
		state.S2.Run = !state.S2.Run
	}
	mutex.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "role", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

var msgHandler mqtt.MessageHandler = func(c mqtt.Client, m mqtt.Message) {
	p := string(m.Payload())
	t := m.Topic()
	mutex.Lock()
	switch t {
	case "home/sensor1/temp": state.S1.Temp = p
	case "home/sensor1/humid": state.S1.Humid = p
	case "home/sensor2/temp": state.S2.Temp = p
	case "home/sensor2/humid": state.S2.Humid = p
	}
	mutex.Unlock()
}

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("go_dashboard")
	mqttClient = mqtt.NewClient(opts)
	mqttClient.Connect().Wait()
	mqttClient.Subscribe("home/sensor1/temp", 0, msgHandler)
	mqttClient.Subscribe("home/sensor1/humid", 0, msgHandler)
	mqttClient.Subscribe("home/sensor2/temp", 0, msgHandler)
	mqttClient.Subscribe("home/sensor2/humid", 0, msgHandler)

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/toggle", toggleHandler)
	http.HandleFunc("/", homeHandler)
	
	// THIS LINE WAS MISSING AND CAUSED THE ERROR:
	log.Println("🌍 Web Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

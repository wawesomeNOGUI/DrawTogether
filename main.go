package main

import (
	//"flag"
	//"html/template"
	"log"
	"net/http"
	//"fmt"
	//"sync"
	//"time"
	//	"strings"

	"github.com/gorilla/websocket"
)

var rDraw [][]byte //2D Slice of bytes where each new user can append there
//drawing data through the WS connection
var oldSlice [][]byte
var users int
var cleaned  = make(chan string) //users use this to tell the mem clean to start
var clean  = make(chan string)
//var newUser = make(chan string)
/*
var rChan = make(chan struct{}) //for telling when to send updated drawing data

type new struct {

}

var empty = new{}    //new empty struct
*/
var upgrader = websocket.Upgrader{} // use default options for upgrader

func test2DSliceEquality(a, b [][]byte) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}


	for i := range a {

			if len(a[i]) != len(b[i]) {
				return false
			}

	}

	for i := range a {
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}

	return true

}

/*
func testSliceEquality(a, b []byte) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {

			if a[i] != b[i] {
				return false
			}

	}
	return true

}
*/

func noMemLeakPls(a *[][]byte){

	for{
		if users > 1 {

				for i := 0; i < users; i++{       //-1 cause first read was in select
					<-clean   //wait for users to be ready for cleanup
					log.Println(i)
				}


				for range *a {

					if len(*a) > 5 {
						//delete first entry from slice (pop)
						*a = (*a)[1:]
						//log.Println(len(*a))
					}

				}

				for i := 0; i < users; i++{
					cleaned <- "cleaned"      //tell users we're done
				}
		}
	}



	}




func closeWriter(rChan chan string){
	 rChan <- "close"
}

func removeUser(){
	users = users - 1
	clean <- "cleanup"  //just to make sure the clean function doesn't block
}


func update(c *websocket.Conn, rChan chan string){
	for {

		select{
			case <-rChan:      //for exiting this function when a user exits
				return
				log.Println("exited")
			default:
				clean <- "cleanup"
				//log.Println("sent")
				<-cleaned       //wait for clean to finish

				if test2DSliceEquality(oldSlice, rDraw) == false {

						for i := 0; i < len(rDraw)-1; i++ {
							if len(rDraw)-1 > i {
								err := c.WriteMessage(websocket.TextMessage, rDraw[i]) //write message back to browser
								if err != nil {
									log.Println("write:", err)
									break
								}
							}
					  }

						//make oldSlice
						oldSlice = make([][]byte, len(rDraw))
						for i := range rDraw {
						    oldSlice[i] = make([]byte, len(rDraw[i]))
						    copy(oldSlice[i], rDraw[i])
						}

				}
		}



	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	users = users + 1

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	rChan := make(chan string)

	defer c.Close()
	defer closeWriter(rChan) //tell the websocket writer to close too
	defer removeUser()


	go update(c, rChan)       //write data back to this user

	for {
		_, message, err := c.ReadMessage() //ReadMessage blocks until SDP message received
		if err != nil {
			log.Println("read:", err)
		}

		if string(message[0]) == "{"{
			rDraw = append(rDraw, message) //udate drawing info in slice
		}else{
			log.Println(message)
		}

		//rChan <- empty
		//log.Println(message, msgType)


	}
}


func main() {
	go noMemLeakPls(&rDraw)

	fileServer := http.FileServer(http.Dir("./public"))
	http.HandleFunc("/echo", echo) //this request comes from webrtc.html
	http.Handle("/", fileServer)




	log.Fatal(http.ListenAndServe(":80", nil))

}

/*
var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server,
"Send" to send a message to the server and "Close" to close the connection.
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
*/

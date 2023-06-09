package main
import ("github.com/gin-gonic/gin"
		"github.com/jmoiron/sqlx"
		_ "github.com/lib/pq"
		//"log"
		"golang.org/x/sync/errgroup"
		"net/http"
		"fmt"
		//"strconv"
		)
//const	(db1ConnectionString = "host=127.0.0.1 port=1433 user=postgres password=password dbname=db1 sslmode=disable"
//		db2ConnectionString = "host=127.0.0.1 port=1433 user=postgres password=password dbname=db2 sslmode=disable"
//		db3ConnectionString = "host=127.0.0.1 port=1433 user=postgres password=password dbname=db3 sslmode=disable"
//		)

var		(g errgroup.Group)

var db1 *sqlx.DB
var db2 *sqlx.DB
var db3 *sqlx.DB

type info struct {
	Name     string `json:"name"`
	Username int `json:"username"`
	Password string `json:"password"`
}

func main(){
	/*
	db1, err := sqlx.Connect("postgres", db1ConnectionString)
	if err != nil {
		log.Fatalln(err)
	}
	defer db1.Close()

	db2, err := sqlx.Connect("postgres", db2ConnectionString)
	if err != nil {
		log.Fatalln(err)
	}
	defer db2.Close()

	db3, err := sqlx.Connect("postgres", db2ConnectionString)
	if err != nil{
		log.Fatalln(err)
	}
	defer db3.Close()
	*/







	//loadRing()
	//testSpeed()










	var coordinator *Coordinator
	//coordinator := &Coordinator{DB: db1}
	//participant1 := &Participant{DB: db1, ID:1, Password:"one"}
	//participant2 := &Participant{DB: db2, ID:2, Password:"two"}
	//participant3 := &Participant{DB: db3, ID:3, Password:"three"}
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	//curl -X POST -H "Content-Type: application/json" -d '{"name": "user1", "username": 1, "password": "one"}' http://localhost:8080/enter

	//type User struct {
	//	Username string
	//	Password string
	//}

	router.POST("/enter", func(c *gin.Context){
		var userData info
		err := c.BindJSON(&userData)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userData.Name = "user1"
		userData.Username = 1
		userData.Password = "one"
		//fmt.Println("Userdata.username", userData.name)
		//fmt.Println("Userdata.username", userData.username)
		//fmt.Println("Userdata.username", userData.password)

		userCreation(userData.Name, userData.Username, userData.Password)

		//fmt.Println("node id is", allnodes[0].ID)
		//fmt.Println("node own public key is", allnodes[0].OwnPublicKey)
		//fmt.Println("leader is", allnodes[0].leader)
		//fmt.Println("leader public key is", allnodes[0].LeadersPublicKey)

		//var stringid := strconv.Itoa(node.ID)
		//var participantname = "participant%s" 
		
		//var p *Participant
		
		//fmt.Println("before election")
		var leaderNode *Node
		leaderElection := NewLeaderElection(allnodes)
		//fmt.Println("past new leader election")
		leaderID, leaderNode, err := leaderElection.ElectLeader()
		//fmt.Println("past elect leader")
		if err != nil {
			fmt.Println("Error electing leader:", err)
			return
		}

		fmt.Println("Leader elected:", leaderID)
		var leaderdb string = leaderNode.database
		//fmt.Println("leader database", leaderdb)

		if leaderdb == "db1"{
			coordinator = &Coordinator{DB: db1}
		}else if leaderdb == "db2" {
			coordinator = &Coordinator{DB: db2}
		}else if leaderdb == "db3" {
			coordinator = &Coordinator{DB: db3}
		}else{
			//fmt.Println("not enough databases")
			return
		}

		c.JSON(200, gin.H{
			"message": "successfully entered network",
		})

		//fmt.Println("Number of participants", len(namesofParticipants))

		register(router, coordinator, namesofParticipants)
	})
	//router.GET("/signup", func(c *gin.Context){
		
	//})

	//register(router, coordinator, namesofParticipants)

	router.Run(":8080")
}

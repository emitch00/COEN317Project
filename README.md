# COEN379Project

install go: https://go.dev/doc/install

setting up module:
go mod init example.com/COEN379Project

Downloading:
go get github.com/gin-gonic/gin
- we are using the gin framework so in the code we import it
go get github.com/jmoiron/sqlx
- we are using sqlx library for database handling
- see more information: https://jmoiron.github.io/sqlx/
go get github.com/lib/pq
- PostgreSQL database driver

Use PostgreSQL:
Set up at least two databases with two tables each...
- name the databases db1 and db2
- name the tables transactions (id, user_id, amount, receiver_id) and wallet (id, funds) where both tables use id as the primary key that autoincrements after every entry (Example: nextval('wallet_id_seq'::regclass))
Set up test user with credentials listed

Run:
- run on terminal: go run main.go twopc.go chord.go leaderelection.go
- on browser: http://localhost:8080/ping -> should result in pong message output
- on terminal there are three actions available...
    1. Enter Network: This will make a node for the particpant that will then allow them to commit further actions.
        curl -X POST -H "Content-Type: application/json" -d '{"Name": "user2", "Username": 2, "Password": "two"}' http://localhost:8080/enter
    2. Check Funds: This will return the current funds in a particpants wallet
        curl -X POST -H "Content-Type: application/json" -d '{"user_id": 1, "password": "one"}' http://localhost:8080/check-funds
        {"funds":46,"status":"success"}
    3. Send Money: This will send money from one node to another. The transaction is recorded in the transactions table and the wallets of each node is updated respectively
        curl -X POST -H "Content-Type: application/json" -d '{"user_id": 1, "password": "one", "receiver_id": 2, "amount": 1}' http://localhost:8080/send-money 

package main
import ("context"
		"github.com/gin-gonic/gin"
		"github.com/jmoiron/sqlx"
		"net/http"
		"fmt"
		)



//
//curl -X POST -H "Content-Type: application/json" -d '{"user_id": 1, "password": "one", "receiver_id": 2, "amount": 100}' http://localhost:8080/send-money
//
//
//curl -X POST -H "Content-Type: application/json" -d '{"user_id": 1, "password": "one"}' http://localhost:8080/check-funds
//

//Define structs for JSON data

type Send struct{
	User int `json:"user_id"`
	Password string `json:"password"`
	Receiver int `json:"receiver_id"`
	//Product string `json:"product_name"`
	//Price int `json:"amount"`
	Amount int `json:"amount"`
}

type Check struct{
	User int `json:"user_id"`
	Password string `json:"password"`
}

//Coordinator data structure
type Coordinator struct{
	DB *sqlx.DB
}
//Participant data structure
type Participant struct{
	DB *sqlx.DB
	ID int
	Password string
}

//Prepare transaction
func (p *Participant) PrepareSendTran(ctx context.Context, tx *sqlx.Tx, prepareData Send) error {
	//Currently only accepts entries into send
	fmt.Println(prepareData.User)
	fmt.Println(prepareData.Amount)
	fmt.Println(prepareData.Receiver)
	//result, err := p.DB.Exec("INSERT INTO orders (user_id, product_name, amount) VALUES ($1, $2, $3)", prepareData.User, prepareData.Product, prepareData.Amount)
	result, err := p.DB.Exec("INSERT INTO transactions (user_id, receiver_id, amount) VALUES ($1, $2, $3)", prepareData.User, prepareData.Receiver, prepareData.Amount)
	
	//alter table .... we want to alter a record for the amount of money they have now after transaction

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	fmt.Println("inserted", rowsAffected, "rows for transaction")

	return nil
}

func (p *Participant) PrepareSendDel(ctx context.Context, tx *sqlx.Tx, prepareData Send) error {
	//Currently only accepts entries into send
	fmt.Println(prepareData.User)
	fmt.Println(prepareData.Amount)
	fmt.Println(prepareData.Receiver)

	var negative int = -(prepareData.Amount)
	//result, err := p.DB.Exec("INSERT INTO orders (user_id, product_name, amount) VALUES ($1, $2, $3)", prepareData.User, prepareData.Product, prepareData.Amount)
	result, err := p.DB.Exec("INSERT INTO wallet (funds) VALUES ($1)", negative)
	
	//alter table .... we want to alter a record for the amount of money they have now after transaction

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	fmt.Println("inserted", rowsAffected, "rows to sender")

	return nil
}

func (p *Participant) PrepareSendIns(ctx context.Context, tx *sqlx.Tx, prepareData Send) error {
	//Currently only accepts entries into send
	fmt.Println(prepareData.User)
	fmt.Println(prepareData.Amount)
	fmt.Println(prepareData.Receiver)
	//result, err := p.DB.Exec("INSERT INTO orders (user_id, product_name, amount) VALUES ($1, $2, $3)", prepareData.User, prepareData.Product, prepareData.Amount)
	result, err := p.DB.Exec("INSERT INTO wallet (funds) VALUES ($1)", prepareData.Amount)
	
	//alter table .... we want to alter a record for the amount of money they have now after transaction

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	fmt.Println("inserted", rowsAffected, "rows to receiver")

	return nil
}

//Prepare transaction
func (p *Participant) PrepareCheck(ctx context.Context, tx *sqlx.Tx, prepareData Check) (int, error) {
	//Currently only accepts entries into send
	//fmt.Println(prepareData.User)
	var funds int
	//var id int
	//result, err := p.DB.Exec("INSERT INTO orders (user_id, product_name, amount) VALUES ($1, $2, $3)", prepareData.User, prepareData.Product, prepareData.Amount)
	err := p.DB.QueryRow("SELECT SUM(funds) FROM wallet").Scan(&funds)
	//show the amount of money they have now

	if err != nil {
		return 0, err
	}

	//rowsAffected, err := result.RowsAffected()
	//if err != nil {
	//	return err
	//}

	//fmt.Println("inserted", rowsAffected, "rows")
	fmt.Println("funds in account: ", funds)
	//fmt.Println("id of row", id)

	if(funds <= 0){
		funds = 0
	}
	return funds, nil
}

//tx is transaction object
//ctx is context of transaction (GO specific)

//Commit transaction
func (p *Participant) Commit(ctx context.Context, tx *sqlx.Tx) error{
	err := tx.Commit()
	if err != nil{
		return err
	}
	return nil
}

//Rollback transaction
func (p *Participant) Rollback(ctx context.Context, tx *sqlx.Tx) error {
	err := tx.Rollback()
	if err != nil {
		return err
	}
	return nil
}

//Execute transaction using 2PC
func (c *Coordinator) ExecuteSend(ctx context.Context, prepareData Send, participants []*Participant) error {

	var sendercheck bool = false
	var receivercheck bool = false
	var participantList []*Participant
	var senderparticipant *Participant
	for _, p := range participants{
		if(prepareData.User == p.ID){
			sendercheck = true
			participantList = append(participantList, p)
			senderparticipant = p
		}
		if(prepareData.Receiver == p.ID){
			receivercheck = true
			participantList = append(participantList, p)
		}
	}

	if(sendercheck != true){
		return fmt.Errorf("user (sender) not defined")
	}

	if(receivercheck != true){
		return fmt.Errorf("user (receiver) not defined")
	}

	fmt.Println("made it past sender and receiver checks")

	//check password
	var passwordcheck bool = false
	if(prepareData.Password == senderparticipant.Password){
		passwordcheck = true
	}

	if(passwordcheck != true){
		return fmt.Errorf("incorrect password")
	}	

	fmt.Println("made it past password check")

	var funds int
	//var id int

	//err := senderparticipant.DB.QueryRow("SELECT funds, MAX(id) FROM wallet GROUP BY funds").Scan(&funds, &id)
	err := senderparticipant.DB.QueryRow("SELECT SUM(funds) FROM wallet").Scan(&funds)
	//show the amount of money they have now

	if err != nil {
		return err
	}

	if (funds < prepareData.Amount){
		return fmt.Errorf("insufficent funds for transaction")
	}

	fmt.Println("made it past funds check")
	//Phase 1: Prepare
	var preparedTransactions []*sqlx.Tx
	var preparedTransactionsSend *sqlx.Tx
	var preparedTransactionsReceiv *sqlx.Tx
	var tx2 *sqlx.Tx
	var tx3 *sqlx.Tx
	for _, p := range participantList {
		//p := participant
		//Beginx is a transaction handle, returns sqlx.Tx
		tx, err := p.DB.Beginx()
		fmt.Println("sqlx.tx object for transaction", tx)
		if err != nil{
			return err
		}
		//prepare database, returns sqlx.Stmt
		err = p.PrepareSendTran(ctx, tx, prepareData)
		if err != nil{
			return err
		}

		if(prepareData.User == p.ID){
			fmt.Println("initiating transaction for sender")
			//tx2, err2 := p.DB.Beginx()
			//fmt.Println("sqlx.tx object for sender", tx2)
			//if err2 != nil{
			//	return err2
			//}

			err2 := p.PrepareSendDel(ctx, tx, prepareData)
			if err2 != nil{
				return err2
			}
		}

		if(prepareData.Receiver == p.ID){
			//fmt.Println("initiating transaction for receiver")
			//tx3, err3 := p.DB.Beginx()
			//fmt.Println("sqlx.tx object for receiver", tx3)
			//if err3 != nil{
			//	return err3
			//}

			//err3 = p.PrepareSendIns(ctx, tx3, prepareData)
			err3 := p.PrepareSendIns(ctx, tx, prepareData)
			if err3 != nil{
				return err3
			}
		}

		//add transaction to prepared list
		preparedTransactions = append(preparedTransactions, tx)
		
		preparedTransactionsSend = tx2
		
		preparedTransactionsReceiv = tx3
		
	}
	fmt.Println("length of prepared transactions: ", preparedTransactions[0])
	fmt.Println("length of prepared transactions: ", preparedTransactions[1])
	fmt.Println("length of send transactions: ", preparedTransactionsSend)
	fmt.Println("length of receive transactions: ", preparedTransactionsReceiv)

	//err1 = senderparticipant.PrepareSendDel(ctx, tx, prepareData)
	//if err1 != nil{
	//	return err1
	//}



	//Phase 2: Commit
	for i, p := range participantList {
		//commit list of transactions
		err := p.Commit(ctx, preparedTransactions[i])
		fmt.Println("length of prepared transactions: ", len(preparedTransactions))
		fmt.Println("length of send transactions: ", preparedTransactionsSend)
		fmt.Println("length of send transactions: ", tx2)
		fmt.Println("length of receiv transactions: ", preparedTransactionsReceiv)
		fmt.Println("length of receiv transactions: ", tx3)
		fmt.Println("committing transaction: ", i)
		if err != nil{
			//Rollback transactions if commit fails
			for _, tx := range preparedTransactions {
				_ = p.Rollback(ctx, tx)
			}
			fmt.Println("returning error...")
			return err
		}

		if(prepareData.User == p.ID){
			fmt.Println("committing sender")
			//err := p.Commit(ctx, preparedTransactionsSend)
			//if err != nil{
				//Rollback transactions if commit fails
				//for _, tx := range preparedTransactionsSend {
				//	_ = p.Rollback(ctx, tx)
				//_ = p.Rollback(ctx, preparedTransactionsSend)
				//}
				//return err
		}

		if(prepareData.Receiver == p.ID){
			fmt.Println("committing receiver")
			//err := p.Commit(ctx, preparedTransactionsReceiv)
			//if err != nil{
				//Rollback transactions if commit fails
				//for _, tx := range preparedTransactionsSend {
				//	_ = p.Rollback(ctx, tx)
				//_ = p.Rollback(ctx, preparedTransactionsReceiv)
				//}
				//return err
		}

	}

	return nil
}

//Execute transaction using 2PC
func (c *Coordinator) ExecuteCheck(ctx context.Context, prepareData Check, participants []*Participant) (int, error) {
	//Phase 1: Prepare
	var preparedTransactions []*sqlx.Tx
	var funds int

	//fmt.Println(prepareData.User)

	//if prepareData.User >= len(participants){
	//	return 0, fmt.Errorf("user not defined")
	//}

	var check bool = false
	var participant *Participant
	for _, p := range participants{
		if(prepareData.User == p.ID){
			check = true
			participant = p
		}
	}

	if(check != true){
		return 0, fmt.Errorf("user not defined")
	}
	
	//var participant *Participant = participants[prepareData.User]

	//check password
	var passwordcheck bool = false
	if(prepareData.Password == participant.Password){
		passwordcheck = true
	}

	if(passwordcheck != true){
		return 0, fmt.Errorf("incorrect password")
	}	
	//for _, p := participant {
		p := participant
		//Beginx is a transaction handle, returns sqlx.Tx
		tx, err := p.DB.Beginx()
		if err != nil{
			return 0, err
		}
		//prepare database, returns sqlx.Stmt
		funds, err = p.PrepareCheck(ctx, tx, prepareData)
		if err != nil{
			return 0, err
		}

		//add transaction to prepared list
		preparedTransactions = append(preparedTransactions, tx)
	//}

	//Phase 2: Commit
	//for i, p := range participants {
		//commit list of transactions
		err = p.Commit(ctx, preparedTransactions[0])
		if err != nil{
			//Rollback transactions if commit fails
			for _, tx := range preparedTransactions {
				_ = p.Rollback(ctx, tx)
			}
			return 0, err
		}
	//}

	return funds, nil
}

//register endpoints
func register(router *gin.Engine, coordinator *Coordinator, participants []*Participant) {
	router.POST("/send-money", func(c *gin.Context) {
		//Prepare data from the request
		var prepareData Send
		err := c.BindJSON(&prepareData)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//var password string
		//c.JSON(http.StatusContinue, gin.H{"enter password":})

		//if(prepareData.Password == p.ID){
		//	sendercheck = true
		//	participantList = append(participantList, p)
		//	senderparticipant = p
		//}

		//Execute distributed transaction
		err = coordinator.ExecuteSend(c, prepareData, participants)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.POST("/check-funds", func(c *gin.Context) {
		//Prepare data from the request
		var prepareData Check
		var funds int
		err := c.BindJSON(&prepareData)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//Execute distributed transaction
		funds, err = coordinator.ExecuteCheck(c, prepareData, participants)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "funds": funds})
	})
}
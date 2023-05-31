package main
import ("context"
		"github.com/gin-gonic/gin"
		"github.com/jmoiron/sqlx"
		"net/http"
		"fmt"
		)


//Define structs for JSON data

type Order struct{
	User int `json:"user_id"`
	Product string `json:"product_name"`
	Price int `json:"amount"`
}

//Coordinator data structure
type Coordinator struct{
	DB *sqlx.DB
}
//Participant data structure
type Participant struct{
	DB *sqlx.DB
}

//Prepare transaction
func (p *Participant) Prepare(ctx context.Context, tx *sqlx.Tx, prepareData Order) error {
	//Currently only accepts entries into orders
	fmt.Println(prepareData.User)
	fmt.Println(prepareData.Price)
	fmt.Println(prepareData.Product)
	result, err := p.DB.Exec("INSERT INTO orders (user_id, product_name, amount) VALUES ($1, $2, $3)", prepareData.User, prepareData.Product, prepareData.Price)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	fmt.Println("inserted", rowsAffected, "rows")

	return nil
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
func (c *Coordinator) Execute(ctx context.Context, prepareData Order, participants []*Participant) error {
	//Phase 1: Prepare
	var preparedTransactions []*sqlx.Tx
	for _, p := range participants {
		//Beginx is a transaction handle, returns sqlx.Tx
		tx, err := p.DB.Beginx()
		if err != nil{
			return err
		}
		//prepare database, returns sqlx.Stmt
		err = p.Prepare(ctx, tx, prepareData)
		if err != nil{
			return err
		}

		//add transaction to prepared list
		preparedTransactions = append(preparedTransactions, tx)
	}

	//Phase 2: Commit
	for i, p := range participants {
		//commit list of transactions
		err := p.Commit(ctx, preparedTransactions[i])
		if err != nil{
			//Rollback transactions if commit fails
			for _, tx := range preparedTransactions {
				_ = p.Rollback(ctx, tx)
			}
			return err
		}
	}

	return nil
}

//register endpoints
func register(router *gin.Engine, coordinator *Coordinator, participants []*Participant) {
	router.POST("/distributed-transaction", func(c *gin.Context) {
		//Prepare data from the request
		var prepareData Order
		err := c.BindJSON(&prepareData)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//Execute distributed transaction
		err = coordinator.Execute(c, prepareData, participants)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
}
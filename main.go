package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


	func connectToMongoDB() (*mongo.Client, error) {
		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			return nil, err
		}
	
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			return nil, err
		}
	
		fmt.Println("Successfully connected to MongoDB!")
		return client, nil
	}
	

	// Function to create the product table in PostgreSQL
// Function to save product data to PostgreSQL
func saveToMongoDB(client *mongo.Client, product Product) error {
    collection := client.Database("your_database_name").Collection("products")
    _, err := collection.ReplaceOne(context.TODO(), bson.M{"name": product.Name}, product, options.Replace().SetUpsert(true))
    return err
}



	type Product struct {
		Name          string
		Price         string
		Variation     float64
		HighestPrice  float64
		LowestPrice   float64
	}
	
	// Map to store products by name
	var productMap map[string]*Product

	


	func main() {
		productMap = make(map[string]*Product)
		

		mongoClient, err := connectToMongoDB()
if err != nil {
    log.Fatal(err)
}
defer mongoClient.Disconnect(context.TODO())




		url := "https://www.amazon.com.br/Monitor-Philips-21-5-com-HDMI/dp/B09BG8BXCK/ref=d_pd_sbs_sccl_3_3/135-0951879-7592106?content-id=amzn1.sym.ec300dba-f3bc-4b4b-a130-7de1ff98f079&pd_rd_i=B09BG8BXCK&psc=1"
		
		


		priceFloat, productName, err := getPriceAndProductName(url)
		if err != nil {
			log.Fatal(err)
		}
	
		fmt.Println("Product Name:", productName)
		fmt.Println("Price:", priceFloat)
	
		exists, err := isProductExistInCSV(productName)
		if err != nil {
			log.Fatal(err)
		}
	
		var previousPrice float64
	
		if exists {
			fmt.Println("Product", productName, "already exists in the CSV.")
			previousPrice, _, err = getPriceAndProductName(url) // Ignore o segundo valor retornado
			if err != nil {
				log.Fatal(err)
			}
		}
	
		currentPrice := previousPrice // Assume que o preço atual é o mesmo que o anterior
	
		variation, highestPrice, lowestPrice, err := calculatePriceVariation(productName, currentPrice)
		if err != nil {
			log.Fatal(err)
		}
	
		fmt.Printf("Price variation: %.2f%%\n", variation)
		updateOrAddProduct(productName, strconv.FormatFloat(priceFloat, 'f', -1, 64), variation, currentPrice, previousPrice)
		err = saveToCsv(productName, strconv.FormatFloat(priceFloat, 'f', -1, 64), variation, highestPrice, lowestPrice)
		if err != nil {
			log.Fatal(err)
		}

		// Replace saveToCsv with saveToMongoDB
		err = saveToMongoDB(mongoClient, Product{
			Name:         productName,
			Price:        strconv.FormatFloat(priceFloat, 'f', -1, 64),
			Variation:    variation,
			HighestPrice: highestPrice,
			LowestPrice:  lowestPrice,
		})
		if err != nil {
			log.Fatal(err)
		}
	
	}
	


	func getPriceAndProductName(url string) (float64, string, error) {
		// Fazemos a requisição HTTP para a página
		response, err := http.Get(url)
		if err != nil {
			return 0.0, "", err
		}
		
		defer response.Body.Close()

		// Parseamos o HTML da página
		document, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return 0.0, "", err
		}

		// Encontramos o elemento que contém o preço
		priceSpan := document.Find("span.a-offscreen")
		price := priceSpan.Text()

		// Check if "R$" is present in the price string
		if !strings.Contains(price, "R$") {
			return 0.0, "", fmt.Errorf("price not found")
		}

		// Split the price using "R$" as the delimiter
		splitedPrice := strings.Split(price, "R$")
		if len(splitedPrice) < 2 {
			return 0.0, "", fmt.Errorf("unable to extract price")
		}
		price = strings.TrimSpace(splitedPrice[1])
		priceFloat, err := strconv.ParseFloat(strings.Replace(price, ",", ".", -1), 64)
		if err != nil {
			return 0, "", fmt.Errorf("Falha ao analisar o preço: %v", err)
		}
		nameSpan := document.Find("span.a-size-large.product-title-word-break")
		productName := nameSpan.Text()
		productName = strings.TrimSpace(productName)

		return priceFloat, productName, nil
	}

	func sendMail(from, password, host string, port int, to []string, message []byte) error {
		auth := smtp.PlainAuth("", from, password, host)

		err := smtp.SendMail(fmt.Sprintf("%s:%d", host, port), auth, from, to, message)
		if err != nil {
			return err
		}

		return nil
	}


	func setupEmail(body string) error {
		from := "andremendes0113@gmail.com"
		password := "0294BE2A8D274556D9584EE59D90DEFC0AB6" // Sua senha do Elastic Email
		to := "andremiranda.aluno@unipampa.edu.br"
		subject := "Preços"
		

		msg := "From: " + from + "\n" +
			"To: " + to + "\n" +
			"Subject: " + subject + "\n\n" +
			body

		err := sendMail(from, password, "smtp.elasticemail.com", 2525, []string{to}, []byte(msg))
		if err != nil {
			fmt.Println("Erro ao enviar o email:", err)
		} else {
			fmt.Println("Email enviado com sucesso!")


		}

		return nil
	}

	func saveToCsv(productName, price string, variation, highestPrice, lowestPrice float64) error {
		file, err := os.OpenFile("output.csv", os.O_RDWR|os.O_CREATE, os.ModeAppend)
		if err != nil {
			return err
		}
		defer file.Close()
	
		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			return err
		}
	
		var exists bool
		for i, record := range records {
			if len(record) == 5 && record[0] == productName {
				// Update existing record
				records[i] = []string{productName, price, fmt.Sprintf("%.2f%%", variation), fmt.Sprintf("%.2f", highestPrice), fmt.Sprintf("%.2f", lowestPrice)}
				exists = true
				break
			}
		}
	
		if !exists {
			// Add new record
			records = append(records, []string{productName, price, fmt.Sprintf("%.2f%%", variation), fmt.Sprintf("%.2f", highestPrice), fmt.Sprintf("%.2f", lowestPrice)})
		}
	
		// Write the updated CSV
		file.Seek(0, 0)
		file.Truncate(0)
		writer := csv.NewWriter(file)
		defer writer.Flush()
	
		for _, record := range records {
			err := writer.Write(record)
			if err != nil {
				return err
			}
		}
	
		return nil
	}
	
	


	func isProductExistInCSV(productName string) (bool, error) {
		// Open the CSV file
		file, err := os.Open("output.csv")
		if err != nil {
			return false, err
		}
		defer file.Close()

		// Create a CSV reader
		reader := csv.NewReader(bufio.NewReader(file))

		// Read and check if the product exists in the CSV
		for {
			record, err := reader.Read()
			if err != nil {
				break
			}

			// Check if the productName matches any record in the CSV
			if len(record) > 0 && strings.TrimSpace(record[0]) == productName {
				return true, nil
			}
		}

		return false, nil
	}
	
func calculatePriceVariation(productName string, currentPrice float64) (float64, float64, float64, error) {
		file, err := os.Open("output.csv")
		if err != nil {
			return 0.0, 0.0, 0.0, err
		}
		defer file.Close()
	
		reader := csv.NewReader(bufio.NewReader(file))
	
		var previousPrice, highestPrice, lowestPrice float64
		var found bool
	
		for {
			record, err := reader.Read()
			if err != nil {
				break
			}
	
			if len(record) > 0 && strings.TrimSpace(record[0]) == productName {
				previousPrice, err = strconv.ParseFloat(record[1], 64)
				if err != nil {
					return 0.0, 0.0, 0.0, err
				}
	
				highestPrice, err = strconv.ParseFloat(record[3], 64)
				if err != nil {
					return 0.0, 0.0, 0.0, err
				}
	
				lowestPrice, err = strconv.ParseFloat(record[4], 64)
				if err != nil {
					return 0.0, 0.0, 0.0, err
				}
	
				found = true
				break
			}
		}
	
		if !found {
			fmt.Println("No previous price found for", productName)
			return 0.0, currentPrice, currentPrice, nil
		}
	
		if currentPrice > highestPrice {
			highestPrice = currentPrice
		} else if currentPrice < lowestPrice {
			lowestPrice = currentPrice
		}
	
		return ((currentPrice - previousPrice) / previousPrice) * 100, highestPrice, lowestPrice, nil
	}
	


	func updateOrAddProduct(productName string, price string, variation float64, currentPrice float64, previousPrice float64) {
		existingProduct, exists := productMap[productName]
		if exists {
			// Update product attributes if necessary
			if currentPrice > existingProduct.HighestPrice {
				existingProduct.HighestPrice = currentPrice
			} else if currentPrice < existingProduct.LowestPrice {
				existingProduct.LowestPrice = currentPrice
			}
			// Recalculate variation
			existingProduct.Variation = ((currentPrice - previousPrice) / previousPrice) * 100
	
			fmt.Printf("Product %s already exists. Updated attributes.\n", productName)
		} else {
			// Create a new product
			productMap[productName] = &Product{
				Name:         productName,
				Price:        price,
				Variation:    variation,
				HighestPrice: currentPrice,
				LowestPrice:  currentPrice,
			}
	
			fmt.Printf("New product added: %s\n", productName)
		}
	}
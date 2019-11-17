// Команда trysome пытается трудоустроить своего разработчика
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	wantedString = "Go" // строка для поиска в телах полученных ответов
	maxQueueLen  = 5    // предельная длина очереди обработки
)

func init() {
	// Направляем вывод сообщений об ошибках в стандартный поток
	log.SetOutput(os.Stderr)
}

func main() {
	wantedBytes := []byte(wantedString) // срез байтов для поиска
	if len(wantedBytes) == 0 {
		log.Fatalf("Have nothing to count")
	}

	queue := make(chan bool, maxQueueLen) // очередь подпрограмм-обработчиков

	total := 0 // сумма найденных вхождений заданной строки

	// Читаем ввод построчно
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		// Ждем места в очереди
		queue <- true

		// Запускаем обработку
		go func(queue chan bool, url string) {
			// Отправляем запрос GET
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Can't issue a GET to %q: %s", url, err.Error())
				return
			}
			defer resp.Body.Close()

			// Пропускаем неудачные ответы
			if resp.StatusCode != http.StatusOK {
				log.Printf("%q replied %s, skiped", url, resp.Status)
				return
			}

			// Вычитываем тело ответа
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Can't read from the %q response body: %s", url, err.Error())
				return
			}

			// Ищем вхождения заданной строки, подсчитываем количество и сумму
			count := bytes.Count(b, wantedBytes) // количество вхождений заданной строки
			total += count

			fmt.Printf("Count for %q: %d\n", url, count)

			// Освобождаем место в очереди
			<-queue

		}(queue, s.Text())
	}
	if err := s.Err(); err != nil {
		log.Fatalf("Can't read a line of the standard input: %s", err.Error())
	}

	fmt.Printf("Total: %d\n", total)
}

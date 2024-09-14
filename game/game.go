package game

import (
	"math/rand"
	"time"

	"github.com/FiveEightyEight/mwfapi/models"
)

func GenerateGameProblems(config models.GameConfig) []models.GameProblem {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	problems := make([]models.GameProblem, 10)

	for i := 0; i < 10; i++ {
		method := config.Methods[rand.Intn(len(config.Methods))]
		num1 := random.Intn(config.Range.Max-config.Range.Min+1) + config.Range.Min
		num2 := random.Intn(config.Range.Max-config.Range.Min+1) + config.Range.Min

		if num2 > num1 {
			num1, num2 = num2, num1
		}

		var answer int
		switch method {
		case models.GameConfigMethodAdd:
			answer = num1 + num2
		case models.GameConfigMethodSubtract:
			answer = num1 - num2
		case models.GameConfigMethodMultiply:
			answer = num1 * num2
		case models.GameConfigMethodDivide:
			if num2 != 0 {
				answer = num1 / num2
			} else {
				num2 = 1
				answer = num1
			}
		}

		problems[i] = models.GameProblem{
			Number1: num1,
			Number2: num2,
			Method:  method,
			Answer:  answer,
		}
	}

	return problems
}

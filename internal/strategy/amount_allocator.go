package strategy

import "math"

func floorToCents(amount float64) float64 {
	return math.Floor((amount+1e-9)*100) / 100
}

// buildOrderAmounts 盡量平均分配全部資金，並確保每筆金額介於 MIN_LOAN 與 MAX_LOAN 之間。
func buildOrderAmounts(totalFunds float64, requestedSplits int, minLoan float64, maxLoan float64) []float64 {
	if requestedSplits <= 0 || totalFunds < minLoan {
		return nil
	}

	orderCount := requestedSplits

	avgAmount := totalFunds / float64(orderCount)

	// 單筆不足最小金額時，減少筆數直到可行。
	for avgAmount < minLoan && orderCount > 1 {
		orderCount--
		avgAmount = totalFunds / float64(orderCount)
	}

	if avgAmount < minLoan {
		return nil
	}

	// 單筆超過最大金額時，維持原筆數並將每筆封頂，剩餘金額保留不下。
	if maxLoan > 0 && avgAmount > maxLoan {
		amounts := make([]float64, 0, orderCount)
		for i := 0; i < orderCount; i++ {
			amounts = append(amounts, maxLoan)
		}
		return amounts
	}

	totalCents := int(math.Floor((totalFunds + 1e-9) * 100))
	baseCents := totalCents / orderCount
	remainderCents := totalCents % orderCount

	amounts := make([]float64, 0, orderCount)
	for i := 0; i < orderCount; i++ {
		cents := baseCents
		if i < remainderCents {
			cents++
		}

		amount := float64(cents) / 100
		if amount < minLoan {
			return nil
		}
		if maxLoan > 0 && amount > maxLoan {
			return nil
		}
		amounts = append(amounts, amount)
	}

	return amounts
}

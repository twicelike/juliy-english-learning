package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type OxfordWordJSON struct {
	Word  string `json:"word"`
	Pos   string `json:"pos"`
	Level string `json:"level"`
}

func main() {
	file, err := os.Open("./cmd/seed_wordsets/oxford.txt")
	if err != nil {
		fmt.Println("Error opening oxford.txt:", err)
		return
	}
	defer file.Close()

	var words []OxfordWordJSON
	currentLevel := "A1"
	scanner := bufio.NewScanner(file)

	// Các từ khóa chỉ loại từ
	posKeywords := []string{
		"indefinite article", "auxiliary v.", "modal v.", "number",
		"prep.", "adv.", "adj.", "conj.", "pron.", "det.", "exclam.", "n.", "v.",
		"prep", "adv", "adj", "conj", "pron", "det", "exclam", "n", "v",
	}

	numSuffixRegex := regexp.MustCompile(`\d+$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Bỏ qua dòng bản quyền
		if strings.Contains(line, "© Oxford University Press") {
			continue
		}

		// Đổi level
		if regexp.MustCompile(`^[A-C][1-2]$`).MatchString(line) {
			currentLevel = line
			continue
		}

		// Tách từ và loại từ (POS)
		wordPart := line
		posPart := ""

		// Tìm vị trí xuất hiện đầu tiên của một từ khóa POS (phải là một từ riêng biệt)
		firstPosIdx := len(line)
		for _, kw := range posKeywords {
			// Tìm kw dưới dạng word boundary
			re := regexp.MustCompile(`\b` + strings.ReplaceAll(kw, ".", `\.`) + `(\b|/|,)`)
			loc := re.FindStringIndex(line)
			if loc != nil && loc[0] > 0 && loc[0] < firstPosIdx {
				firstPosIdx = loc[0]
			}
		}

		if firstPosIdx < len(line) {
			wordPart = line[:firstPosIdx]
			posPart = line[firstPosIdx:]
		}

		// Xóa nội dung trong ngoặc đơn ở word (VD: "bank (money)")
		if idx := strings.Index(wordPart, "("); idx != -1 {
			wordPart = wordPart[:idx]
		}

		// Chỉ lấy POS đầu tiên nếu có nhiều POS (phân cách bởi , hoặc /)
		firstPos := posPart
		if idx := strings.IndexAny(posPart, ",/"); idx != -1 {
			firstPos = posPart[:idx]
		}
		firstPos = strings.TrimSpace(firstPos)

		// Có những dòng có 2 từ như "a, an" -> tách ra
		subWords := strings.Split(wordPart, ",")
		for _, w := range subWords {
			w = strings.TrimSpace(w)
			// Xóa số ở cuối chữ (VD: can1 -> can)
			w = numSuffixRegex.ReplaceAllString(w, "")

			if w != "" {
				words = append(words, OxfordWordJSON{
					Word:  w,
					Pos:   firstPos,
					Level: currentLevel,
				})
			}
		}
	}

	outFile, err := os.Create("./cmd/seed_wordsets/oxford.json")
	if err != nil {
		fmt.Println("Error creating json:", err)
		return
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	encoder.Encode(words)

	fmt.Printf("✅ Đã parse thành công %d từ và lưu vào oxford.json\n", len(words))
}

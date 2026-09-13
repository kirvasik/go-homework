package main

import "fmt"

func twoSum(nums []int, target int) []int {
	seen := make(map[int]int, len(nums))
	for i, v := range nums {
		if j, ok := seen[target-v]; ok {
			return []int{j, i}
		}
		seen[v] = i
	}
	return nil
}

func main() {
	nums := []int{2, 7, 11, 15}
	target := 9

	result := twoSum(nums, target)
	fmt.Printf("twoSum(%v, %d) = %v\n", nums, target, result)
}

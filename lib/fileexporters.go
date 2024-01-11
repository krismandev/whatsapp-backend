package lib

import (
	"os"
	"encoding/csv"
)

//ExportToCSV is use to export data to string [row][coloumn]
func ExportToCSV(filename string,data [][]string) (err error) {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    writer := csv.NewWriter(file)
    defer writer.Flush()
    writer.WriteAll(data)
    // for _, value := range data {
    //     if err := writer.Writeln(value); err != nil {
    //         return err 
    //     }
    // }
    return nil
}
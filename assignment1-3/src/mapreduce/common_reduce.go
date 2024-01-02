package mapreduce

import (
	"encoding/json"
	"os"
	"sort" // to sort the keys in klist
)

// doReduce does the job of a reduce worker: it reads the intermediate
// key/value pairs (produced by the map phase) for this task, sorts the
// intermediate key/value pairs by key, calls the user-defined reduce function
// (reduceF) for each key, and writes the output to disk.
func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	// TODO:
	// You will need to write this function.
	// You can find the intermediate file for this reduce task from map task number
	// m using reduceName(jobName, m, reduceTaskNumber).
	// Remember that you've encoded the values in the intermediate files, so you
	// will need to decode them. If you chose to use JSON, you can read out
	// multiple decoded values by creating a decoder, and then repeatedly calling
	// .Decode() on it until Decode() returns an error.
	//
	// You should write the reduced output in as JSON encoded KeyValue
	// objects to a file named mergeName(jobName, reduceTaskNumber). We require
	// you to use JSON here because that is what the merger than combines the
	// output from all the reduce tasks expects. There is nothing "special" about
	// JSON -- it is just the marshalling format we chose to use. It will look
	// something like this:
	//
	// enc := json.NewEncoder(mergeFile)
	// for key in ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
	//
	// Use checkError to handle errors.

	//output should be stored in the filename that you get after passing the name through mergeName function, not a big deal
	//reduce OP = call reduceF
	//fname := mergeName(jobName, reduceTaskNumber)
	//create a key value list
	//start encoding this (encoder runs here)
	//enc = encode reduce OP
	//put enc in fname

	maps := make(map[string][]string)
	var klist []string
	for m := 0; m < nMap; m++ {
		fname := reduceName(jobName, m, reduceTaskNumber)
		file, _ := os.Open(fname)
		defer file.Close()
		fnew := json.NewDecoder(file)
		//for i := range fnew {
		//
		//}
		for fnew.More() {
			var content KeyValue
			fnew.Decode(&content) // decoding the content's content
			maps[content.Key] = append(maps[content.Key], content.Value)
		}
	}
	for i := range maps {
		klist = append(klist, i)
	}
	sort.Strings(klist)
	opname := mergeName(jobName, reduceTaskNumber)
	opfile, _ := os.OpenFile(opname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	enc := json.NewEncoder(opfile)
	for _, i := range klist {
		opreduce := reduceF(i, maps[i])
		kv := KeyValue{i, opreduce}
		err := enc.Encode(&kv)
		checkError(err)
	}
}

package mapreduce

import (
	"encoding/json"
	"hash/fnv"
	"math"
	//"html/template"
	//"log"
	//"math"
	"os"
	//"reflect"
)

// doMap does the job of a map worker: it reads one of the input files
// (inFile), calls the user-defined map function (mapF) for that file's
// contents, and partitions the output into nReduce intermediate files.
func doMap(
	jobName string, // the name of the MapReduce job
	mapTaskNumber int, // which map task this is
	inFile string,
	nReduce int, // the number of reduce task that will be run ("R" in the paper)
	mapF func(file string, contents string) []KeyValue,
) {
	// TODO:
	// You will need to write this function.
	// You can find the filename for this map task's input to reduce task number
	// r using reduceName(jobName, mapTaskNumber, r). The ihash function (given
	// below doMap) should be used to decide which file a given key belongs into.
	//
	// The intermediate output of a map task is stored in the file
	// system as multiple files whose name indicates which map task produced
	// them, as well as which reduce task they are for. Coming up with a
	// scheme for how to store the key/value pairs on disk can be tricky,
	// especially when taking into account that both keys and values could
	// contain newlines, quotes, and any other character you can think of.
	//
	// One format often used for serializing data to a byte stream that the
	// other end can correctly reconstruct is JSON. You are not required to
	// use JSON, but as the output of the reduce tasks *must* be JSON,
	// familiarizing yourself with it here may prove useful. You can write
	// out a data structure as a JSON string to a file using the commented
	// code below. The corresponding decoding functions can be found in
	// common_reduce.go.
	//
	//   enc := json.NewEncoder(file)
	//   for _, kv := ... {
	//     err := enc.Encode(&kv)
	//
	// Remember to close the file after you have written all the values!
	// Use checkError to handle errors.

	b, err := os.ReadFile(inFile)
	checkError(err)
	s := string(b)
	maps := mapF(inFile, s)            // passing to mapF function
	flist := make(map[string]*os.File) // making a list of files for the filename and the contents
	//use ihash
	for _, i := range maps {
		//j := i.Key
		//index1 := ihash(j)		// hashing to partition
		//index2 := math.Mod(float64(index1), float64(nReduce))  // to calculate the basic hash function (from paper)
		//index := ihash(i.Key) % uint32(nReduce)
		index := math.Mod(float64(ihash(i.Key)), float64(nReduce))
		fname := reduceName(jobName, mapTaskNumber, int(index))

		//flag := 0
		//for _, f := flist[fname] {
		//	if f == fname {
		//		flag = 1
		//	}
		//}
		//if flag == 0 {		// if file not present, add to it
		//	file, _ := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
		//	flist[fname] = file
		//	file.Close()
		//}

		if _, i := flist[fname]; !i {
			file, _ := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
			flist[fname] = file
			defer file.Close()
		}
		enc := json.NewEncoder(flist[fname])
		er := enc.Encode(&i)
		checkError(er)
	}
}

// using math.modulus function for hashing
// give file to fmap and take in the output
//partition the outputs
//

func ihash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

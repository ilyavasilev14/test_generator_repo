package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var exerciseList []Exercise
var trustedExerciseList []int
var secretCode = os.Getenv("TESTGENERATORREPO_SECRET")

type Exercise struct {
    Name        string
    Trusted     bool
    LuaCode     string
}

type ExerciseResponse struct {
    Name    string
    Id      uint64
    LuaCode string
}

type ExerciseListResponse struct {
    Names      []string
    Ids        []uint64
    TrustedIds []uint64
    EndIdx     uint64
}


func handler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path[1:] == "list" {
        currentIdxQuery := r.URL.Query().Get("currentIdx")
        currentIdx, err := strconv.ParseUint(currentIdxQuery, 10, 64)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"%s\" }", "Cannot convert currentIdx " + currentIdxQuery + " to uint.")
            return
        }
        endRange := currentIdx + 50
        if endRange > uint64(len(exerciseList)) {
            endRange = uint64(len(exerciseList))
        }

        responseExerciseList := ExerciseListResponse {}

        for idx,exercise := range exerciseList[currentIdx:endRange] {
            if exercise.Name == "" { break }
            responseExerciseList.Names = append(responseExerciseList.Names, exercise.Name)
            responseExerciseList.Ids = append(responseExerciseList.Ids, currentIdx + uint64(idx))
        }
        responseExerciseList.EndIdx = endRange - 1

        response, err := json.Marshal(responseExerciseList)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Cannot marshal responseExerciseList to json! Err:", err.Error())
            return
        }
        responseString := string(response)
        fmt.Fprint(w, responseString)
    } else if r.URL.Path[1:] == "uploadExercise" {
        data,err := io.ReadAll(r.Body);
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to read body of the response! Err:", err.Error())
            return
        }
        dataString := string(data)
        exerciseData := Exercise {}
        unmarshalErr := json.Unmarshal([]byte(dataString), &exerciseData)
        if unmarshalErr != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Unmarshal error! Err:", unmarshalErr.Error())
            return
        }
        exerciseData.Trusted = false
        exerciseList = append(exerciseList, exerciseData)

        jsonData, err := json.Marshal(exerciseData)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to get marshal exercise! Err:", err.Error())
            return
        }

        fileDirectory, err := os.UserHomeDir()
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to get UserHomeDir! Err:", err.Error())
            return
        }
        fileDirectory += "/TestGeneratorRepo/" + strconv.Itoa(len(exerciseList) - 1) + ".json"

        writeErr := os.WriteFile(fileDirectory, jsonData, 0644)
        if writeErr != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to write to the file. Err:", writeErr.Error())
            return
        }
    } else if r.URL.Path[1:] == "trustedList" {
        currentIdxQuery := r.URL.Query().Get("currentIdx")
        currentIdx, err := strconv.ParseUint(currentIdxQuery, 10, 64)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"%s\" }", "Cannot convert currentIdx " + currentIdxQuery + " to uint.")
            return
        }
        endRange := currentIdx + 50
        if endRange > uint64(len(trustedExerciseList)) {
            endRange = uint64(len(trustedExerciseList))
        }

        responseExerciseList := ExerciseListResponse {}

        for trustedIdx,exerciseIdx := range trustedExerciseList[currentIdx:endRange] {
            exercise := exerciseList[exerciseIdx]
            if exercise.Name == "" { break }
            responseExerciseList.Names = append(responseExerciseList.Names, exercise.Name)
            responseExerciseList.Ids = append(responseExerciseList.Ids, uint64(exerciseIdx))
            responseExerciseList.TrustedIds = append(responseExerciseList.TrustedIds, currentIdx + uint64(trustedIdx))
        }
        responseExerciseList.EndIdx = endRange - 1

        response, err := json.Marshal(responseExerciseList)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Cannot marshal responseExerciseList to json! Err:", err.Error())
            return
        }
        responseString := string(response)
        fmt.Fprint(w, responseString)
    } else if r.URL.Path[1:] == "getExercise" {
        idxQuery := r.URL.Query().Get("idx")
        idx, err := strconv.ParseUint(idxQuery, 10, 64)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"%s\" }", "Cannot convert idx " + idxQuery + " to uint.")
            return
        }
        exercise := exerciseList[idx]

        response, err := json.Marshal(exercise)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Cannot marshal exercise to json! Err:", err.Error())
            return
        }
        responseString := string(response)
        fmt.Fprint(w, responseString)
    } else if r.URL.Path[1:] == "search" {
        query := r.URL.Query().Get("query")
        query = strings.ReplaceAll(query, " ", "")
        query = strings.ToLower(query)

        response := ExerciseListResponse {}
        for idx,exercise := range exerciseList {
            if len(response.Names) > 100 { break }
            name := strings.ReplaceAll(exercise.Name, " ", "")
            name = strings.ToLower(name)
            if strings.Contains(name, query) {
                response.Names = append(response.Names, exercise.Name)
                response.Ids = append(response.Ids, uint64(idx))
            }
        }
        responseStringBytes, err := json.Marshal(response)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Cannot marshal response to json! Err:", err.Error())
            return
        }
        responseString := string(responseStringBytes)
        fmt.Fprint(w, responseString)
    } else if r.URL.Path[1:] == "trustedSearch" {
        query := r.URL.Query().Get("query")
        query = strings.ReplaceAll(query, " ", "")
        query = strings.ToLower(query)

        response := ExerciseListResponse {}
        for idx,exerciseIdx := range trustedExerciseList {
            exercise := exerciseList[exerciseIdx]
            if len(response.Names) > 100 { break }
            name := strings.ReplaceAll(exercise.Name, " ", "")
            name = strings.ToLower(name)
            if strings.Contains(name, query) {
                response.Names = append(response.Names, exercise.Name)
                response.Ids = append(response.Ids, uint64(idx))
            }
        }
        responseStringBytes, err := json.Marshal(response)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Cannot marshal response to json! Err:", err.Error())
            return
        }
        responseString := string(responseStringBytes)
        fmt.Fprint(w, responseString)
    } else if r.URL.Path[1:] == "trustedSearch" {
        query := r.URL.Query().Get("query")
        query = strings.ReplaceAll(query, " ", "")
        query = strings.ToLower(query)

        response := ExerciseListResponse {}
        for idx,exerciseIdx := range trustedExerciseList {
            exercise := exerciseList[exerciseIdx]
            if len(response.Names) > 100 { break }
            name := strings.ReplaceAll(exercise.Name, " ", "")
            name = strings.ToLower(name)
            if strings.Contains(name, query) {
                response.Names = append(response.Names, exercise.Name)
                response.Ids = append(response.Ids, uint64(idx))
            }
        }
        responseStringBytes, err := json.Marshal(response)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Cannot marshal response to json! Err:", err.Error())
            return
        }
        responseString := string(responseStringBytes)
        fmt.Fprint(w, responseString)
    } else if r.URL.Path[1:] == "markTrusted" {
        secret := r.URL.Query().Get("secret")
        strId := r.URL.Query().Get("id")
		if secret != secretCode {
            fmt.Fprintf(w, "Cannot mark the exercise as trusted! The secret values don't match.")
            fmt.Println("Cannot mark the exercise as trusted! The secret values don't match.")
            return
		}
		id, err := strconv.Atoi(strId)
		if err != nil {
            fmt.Fprintf(w, "Cannot mark the exercise as trusted! Failed to convert the exercise ID to int. " + strId)
            fmt.Println("Cannot mark the exercise as trusted! Failed to convert the exercise ID to int.", strId)
            return
		}

		if len(exerciseList) - 1 < id {
            fmt.Fprintf(w, "Cannot mark the exercise as trusted! Failed to get an exercise with ID " + strId)
            fmt.Println("Cannot mark the exercise as trusted! Failed to get an exercise with ID ", strId)
            return
		}

		exercise := exerciseList[id]
		exercise.Trusted = true
		trustedExerciseList = append(trustedExerciseList, id)
		exerciseList[id] = exercise

        jsonData, err := json.Marshal(exercise)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to get marshal exercise! Err:", err.Error())
            return
        }

        fileDirectory, err := os.UserHomeDir()
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to get UserHomeDir! Err:", err.Error())
            return
        }
        fileDirectory += "/TestGeneratorRepo/" + strconv.Itoa(id) + ".json"

        writeErr := os.WriteFile(fileDirectory, jsonData, 0644)
        if writeErr != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to write to the file. Err:", writeErr.Error())
            return
        }
    } else if r.URL.Path[1:] == "markUntrusted" {
        secret := r.URL.Query().Get("secret")
        strId := r.URL.Query().Get("id")
		if secret != secretCode {
            fmt.Fprintf(w, "Cannot mark the exercise as untrusted! The secret values don't match.")
            fmt.Println("Cannot mark the exercise as untrusted! The secret values don't match.")
            return
		}
		id, err := strconv.Atoi(strId)
		if err != nil {
            fmt.Fprintf(w, "Cannot mark the exercise as untrusted! Failed to convert the exercise ID to int. " + strId)
            fmt.Println("Cannot mark the exercise as untrusted! Failed to convert the exercise ID to int.", strId)
            return
		}

		if len(exerciseList) - 1 < id {
            fmt.Fprintf(w, "Cannot mark the exercise as untrusted! Failed to get an exercise with ID " + strId)
            fmt.Println("Cannot mark the exercise as untrusted! Failed to get an exercise with ID ", strId)
            return
		}

		exercise := exerciseList[id]
		exercise.Trusted = false
		exerciseList[id] = exercise

		removeIDs := make([]int, 0)
		for trustedIdx, trustedExercise := range trustedExerciseList {
			if trustedExercise == id {
				removeIDs = append(removeIDs, trustedIdx)
			}
		}
		for _, removeID := range removeIDs {
			slice2 := trustedExerciseList[removeID+1:]
			trustedExerciseList = append(trustedExerciseList[:removeID], slice2...)
		}

        jsonData, err := json.Marshal(exercise)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to get marshal exercise! Err:", err.Error())
            return
        }

        fileDirectory, err := os.UserHomeDir()
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to get UserHomeDir! Err:", err.Error())
            return
        }
        fileDirectory += "/TestGeneratorRepo/" + strconv.Itoa(id) + ".json"

        writeErr := os.WriteFile(fileDirectory, jsonData, 0644)
        if writeErr != nil {
            fmt.Fprintf(w, "{ \"error\": \"Internal error.\" }")
            fmt.Println("Failed to write to the file. Err:", writeErr.Error())
            return
        }
	}
}

func main() {
	if secretCode == "" {
		log.Fatal("Set the secret code for remote server management by setting",
			"the ENVIRONMENT variable 'TESTGENERATORREPO_SECRET'")
	}

    loadExercises()
	fmt.Println("Scan finished!")

    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":4001", nil))
}

func loadExercises() {
    homeDirectory, err := os.UserHomeDir()
    if err != nil {
        fmt.Println("Failed to get UserHomeDir:", err.Error())
        return
    }
    homeDirectory += "/TestGeneratorRepo/"
    fmt.Println("Scanning", homeDirectory, "directory for exercises")

    filepath.WalkDir(homeDirectory, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            fmt.Println("WalkDir error:", err.Error())
            return nil
        }

        if d.Type().IsDir() {
            return nil
        }

        fileContents, err := os.ReadFile(path)
        if err != nil {
            fmt.Println("Failed to read file ", path, ":", err.Error())
            return nil
        }

        exercise := Exercise {}
        jsonErr := json.Unmarshal(fileContents, &exercise)
        if jsonErr != nil {
            fmt.Println("Failed to unmarshal contents of", path, ":", jsonErr.Error())
            return nil
        }

        exerciseList = append(exerciseList, exercise)
        if exercise.Trusted {
            trustedExerciseList = append(trustedExerciseList, len(exerciseList) - 1)
        }

        return nil
    });
}

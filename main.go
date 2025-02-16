package main

import (
	"encoding/json"
	"fmt"
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
        endRange := currentIdx + 2
        if endRange > uint64(len(exerciseList)) {
            endRange = uint64(len(exerciseList))
        }

        responseExerciseList := ExerciseListResponse {}

        for idx,exercise := range exerciseList[currentIdx:endRange] { // 2 to just test
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
    } else if r.URL.Path[1:] == "trustedList" {
        currentIdxQuery := r.URL.Query().Get("currentIdx")
        currentIdx, err := strconv.ParseUint(currentIdxQuery, 10, 64)
        if err != nil {
            fmt.Fprintf(w, "{ \"error\": \"%s\" }", "Cannot convert currentIdx " + currentIdxQuery + " to uint.")
            return
        }
        endRange := currentIdx + 2
        if endRange > uint64(len(trustedExerciseList)) {
            endRange = uint64(len(trustedExerciseList))
        }

        responseExerciseList := ExerciseListResponse {}

        for trustedIdx,exerciseIdx := range trustedExerciseList[currentIdx:endRange] { // 2 to just test
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
    } else if r.URL.Path[1:] == "uploadExercise" {
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
    }
}

func main() {
    loadExercises()

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

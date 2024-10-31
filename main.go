package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "strings"
    "flag"
    "unicode"
    "sort"
    "time"
)

func time_to_str(t time.Time) string {
    return t.Format(time.RFC3339)
}
func str_to_time(str string) time.Time {
    t,err := time.Parse(time.RFC3339, str)
    if err != nil {
        panic(err)
    }
    return t
}


type lineentry struct {
    Count int   `json:"count"`
    Line string `json:"line"`
}
// LineEntries implements sort.Interface for []lineentry
type LineEntries []lineentry
func (l LineEntries) Len() int           { return len(l) }
func (l LineEntries) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l LineEntries) Less(i, j int) bool {
    if l[i].Count < l[j].Count {
        return true
    }
    iRunes := []rune(l[i].Line)
    jRunes := []rune(l[j].Line)

    max := len(iRunes)
    if max > len(jRunes) {
        max = len(jRunes)
    }

    for idx := 0; idx < max; idx++ {
        ir := iRunes[idx]
        jr := jRunes[idx]

        lir := unicode.ToLower(ir)
        ljr := unicode.ToLower(jr)

        if lir != ljr {
            return lir < ljr
        }

        // the lowercase runes are the same, so compare the original
        if ir != jr {
            return ir < jr
        }
    }

    // If the strings are the same up to the length of the shortest string,
    // the shorter string comes first
    return len(iRunes) < len(jRunes)
}

type settings struct {
    Name string `json:"name"`
    Edited string `json:"edited"`
}

type db_state struct {
    Conf settings   `json:"conf"`
    DB []lineentry  `json:"db"`
}
// Create new empty db_state
func db_state_init_empty(name string) *db_state {
    var new_db db_state
    new_db.Conf = settings{name,time_to_str(time.Now())}
    new_db.DB = make([]lineentry, 0)
    return &new_db
}
// Get base path where all dbs are saved
func db_state_get_save_dir() string {
    db_save_path := os.Getenv("XDG_DATA_HOME")
    if len(db_save_path) == 0 {
        db_save_path = "~/.local/share"
    }
    db_save_path = db_save_path + "/go-listcounter"
    return db_save_path
}
// Get full path to the db save file
func db_state_get_save_path(name string) string {
    return db_state_get_save_dir() + "/" + name + ".json"
}
// Load a db file, requires name
func db_state_load(name string) *db_state {
    db_save_path := db_state_get_save_path(name)

    new_db := db_state_init_empty(name)
    // check if file exists
    file, err := os.Open(db_save_path)
    defer file.Close()
    if err != nil {
        // fmt.Fprintln(os.Stderr, "Can't open file, creating a new one")
        return new_db
    }
    fileInfo, err := file.Stat()
    if err != nil {
        // fmt.Fprintln(os.Stderr,"Can't stat file, creating a new one")
        return new_db
    }
    if fileInfo.IsDir(){
        panic("File path is a directory!")
    } else {
        bytes_read := make([]byte, 1024*1024*1) // Up to 1MB of data
        n, err := file.Read(bytes_read)
        if err != nil && err != io.EOF {
            panic(err)
        }
        if n == 0 {
            return new_db
        }
        new_db.from_json(bytes_read[0:n])
        return new_db
    }
}
// Save db to file
func (db *db_state) db_state_save() {
    db_save_path := db_state_get_save_dir()
    err := os.MkdirAll(db_save_path, 0755)
    if err != nil {
        panic(err)
    }
    db_save_path = db_state_get_save_path(db.Conf.Name)
    json_serialization := db.to_json()
    err = os.WriteFile(db_save_path, []byte(json_serialization), 0644)
    if err != nil{
        panic(err)
    }
}
// Drop empty entries from db
func (db *db_state) _drop_empty(){
    i := 0
    for i < len(db.DB){
        if db.DB[i].Count <= 0 {
            db._drop_entry(i)
        } else {
            i++
        }
    }
}
// Drop entry by index
func (db *db_state) _drop_entry(idx int){
    db.DB[idx] = db.DB[len(db.DB)-1] // Move last element forwards
    db.DB = db.DB[:len(db.DB)-1]
}
// Get index of the given entry, else -1
func (db *db_state) _find_entry(line string) int {
    for i:=0; i<len(db.DB); i++ {
        if db.DB[i].Line == line {
            return i
        }
    }
    return -1
}
// Increment counter of an entry, add entry if it doesn't yet exists
func (db *db_state) increment(line string) {
    // find entry
    li := db._find_entry(line)

    if li == -1 {
        db.DB = append(db.DB, lineentry{Count:1,Line:line})
    } else {
        db.DB[li].Count++
    }
}
// Return json serialization string
func (db *db_state) to_json() string {
    db._drop_empty()
    b,err := json.Marshal(db)
    if err != nil {
        fmt.Println(err)
        panic(err)
    }
    return string(b)
}
// Load values from json into a db_state
func (db *db_state) from_json(data []byte) error {
    if err := json.Unmarshal(data, db); err != nil {
        panic(err)
    }
    db._drop_empty()
    return nil
}
// Sort lineentries descending
func (db *db_state) sort_desc() {
    sort.Sort(LineEntries(db.DB))
}
// Sort lineentries ascending
func (db *db_state) sort_asc() {
    sort.Sort(sort.Reverse(LineEntries(db.DB)))
}
func (db *db_state) add_stdin() {
    // TODO: field/delim functionallity

    // Check if stdin is connected
    stat, _ := os.Stdin.Stat()
    if (stat.Mode() & os.ModeCharDevice) == 0 {
        // fmt.Println("data is being piped to stdin")
    } else {
        // fmt.Println("stdin is from a terminal")
        return
    }

    // Read stdin
    reader := bufio.NewReader(os.Stdin)
    text, err := reader.ReadString('\n')
    counter := 0
    for err == nil {
        // counter += 1
        // fmt.Print(text)
        // str := string()
        text = strings.ReplaceAll(text, "\n", "")
        if db._find_entry(text) == -1 {
            db.DB = append(db.DB, lineentry{Count:counter,Line:strings.ReplaceAll(text,"\n", "")})
        }
        text, err = reader.ReadString('\n')
    }
}


func main() {
    // help v delim^field^positional=1 v remove^positional=1 v path^positional=1 v positional=1

    argsHelp := flag.Bool("help", false, "Print this message.")
    argsDelim := flag.String("delimiter", "", "Delimiter to split the line into parts, requires -f.")
    argsField := flag.Int("field", -1, "Position of the ID part of the line, requires -d.")
    argsRemove := flag.String("remove", "", "Remove specified line from database.")
    argsName := flag.String("name", "", "REQUIRED! Name for the database.")
    argsPath := flag.Bool("path", false, "Print path to database save location.")
    argsIncrement := flag.String("inc", "", "Increment the line with the specified identifier.")
    flag.Parse()
    if !( *argsHelp || *argsDelim!=""&&*argsField!=-1&&*argsName!="" || *argsRemove!=""&&*argsName!="" || *argsPath&&*argsName!="" || *argsName!="" ) {
        flag.Usage()
        return
    }
    if *argsDelim!=""&&*argsField==-1 || *argsDelim==""&&*argsField!=-1 {
        fmt.Fprintln(os.Stderr, "--delimiter and --field can only be used together!")
        return
    }
    if *argsHelp {
        flag.Usage()
        return
    }

    if *argsDelim!=""&&*argsField!=-1&&*argsName!="" {
        if *argsField < 0 {
            fmt.Fprintln(os.Stderr, "--field must be >= 0!")
        }

        if *argsIncrement!="" {

        } else {

        }

        fmt.Println("Use delim: name:", *argsName, "delim:", *argsDelim, "field:", *argsField)
        return
    }
    if *argsRemove!=""&&*argsName!="" {
        db := db_state_load(*argsName)
        idx:= db._find_entry(*argsRemove)
        if idx == -1 {
            fmt.Fprintf(os.Stderr, "Entry <%s> not found in DB!\n", *argsRemove)
        } else {
            db._drop_entry(idx)
            db.db_state_save()
        }
        return
    }
    if *argsPath&&*argsName!="" {
        db := db_state_load(*argsName)
        text := db_state_get_save_path(db.Conf.Name)
        fmt.Printf("Path: %s\n",text)
        return
    }
    if *argsName!="" {
        // fmt.Println("New db with name:", *argsName)
        db := db_state_load(*argsName)
        db.add_stdin()
        if *argsIncrement!="" {
            db.increment(*argsIncrement)
            // fmt.Println("Incrementing line:", *argsIncrement)
            db.db_state_save()
        } else {
            db.sort_asc()
            for i := 0; i< len(db.DB); i++ {
                fmt.Printf("%s\n", db.DB[i].Line)
            }
        }
        return
    }







    new_db := db_state_init_empty("testname")
    new_db.DB = append(new_db.DB, lineentry{Count:3,Line:"test-line"})


    fmt.Println("\nAll read!")

    for i := 0; i< len(new_db.DB); i++ {
        fmt.Printf("%d: %s\n", new_db.DB[i].Count, new_db.DB[i].Line)
    }
    fmt.Println("\nNow sorted!")
    new_db.sort_asc()
    for i := 0; i< len(new_db.DB); i++ {
        fmt.Printf("%d: %s\n", new_db.DB[i].Count, new_db.DB[i].Line)
    }

    fmt.Println("\n---")

    // Test mkdir
    ndb := db_state_load("a")
    fmt.Println(ndb.to_json())
    ndb.db_state_save()
    ndb = db_state_load("a")
    fmt.Println(ndb.to_json())
}

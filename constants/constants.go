package constants

import "os"



var PORT = os.Getenv("PORT")
var AuthList []Auth
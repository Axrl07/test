package global

// variables globales
var (
	// -------------------------- autenticación --------------------------

	// UsuarioActual
	UsuarioActual = &Sesion{
		IsLoggedIn:  false,
		Username:    "",
		Password:    "",
		PartitionID: "",
	}

	// -------------------------- discos y montajes --------------------------

	// Lista con todo el abecedario (para darle una letra a cada disco)
	Alfabeto = []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	}

	// Mapa para almacenar la asignación de letras a los diferentes paths (path, letra)
	LetraDelPath = make(map[string]string)

	// Índice para la siguiente letra disponible en el abecedario
	NextLetterIndex = 0

	// almacena particiones montada y el path del disco al que pertenece
	ParticionesMontadas map[string]string = make(map[string]string)

	// almacena información de los discos MOntados
	Montaje map[string]Montadas = make(map[string]Montadas)

	// ruta de discos locales
	RutaDiscosLocales = ""

	TablaHTMLJournaling = ""
)

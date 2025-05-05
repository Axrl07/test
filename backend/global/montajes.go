package global

import (
	"errors"
	"fmt"
)

// GetLetter obtiene la letra asignada a un path
func GetLetra(path string) (string, error) {
	// Asignar una letra al path si no tiene una asignada
	if _, exists := LetraDelPath[path]; !exists {
		if NextLetterIndex < len(Alfabeto) {
			LetraDelPath[path] = Alfabeto[NextLetterIndex]
			NextLetterIndex++
		} else {
			fmt.Println("Error: no hay más letras disponibles para asignar")
			return "", errors.New("Error: no hay más letras disponibles para asignar")
		}
	}

	return LetraDelPath[path], nil
}

// me almacena cuantos discos utilizo y el contador de particiones
type Montadas struct {
	Path     string //Path del Disco
	Letra    string // letra que identifica el disco
	Contador int    //Contador del numero de última particion montada
}

// Ingresa la informacion al Struct
func AgregarMontaje(path string) (*Montadas, error) {

	// verificamos existencia del disco en Montaje
	discoMontado := Montaje[path]

	// si no existe
	if discoMontado.Path == "" {
		// hay que asignarle una letra al path
		nuevaLetra, err := GetLetra(path)
		if err != nil {
			fmt.Println("ERROR: ha ocurrido un error obteniendo la letra:", err)
			return nil, err
		}
		nuevoDisco := Montadas{Path: path, Letra: nuevaLetra, Contador: 1}
		// guardamos bajo el path, la letra, el propio path y el contador
		Montaje[path] = nuevoDisco
		//retornamos el nuevo montaje
		return &nuevoDisco, nil
	}

	// si ya existe lo retornamos
	return &discoMontado, nil
}

// actualiza el montaje
func ActualizarMontaje(path string) {
	discoMontado := Montaje[path]
	discoMontado.Contador += 1
	Montaje[path] = discoMontado
}

// // Lista con todo el abecedario
// var alfabeto = []string{
// 	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
// 	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
// }

// // Mapa para almacenar la asignación de letras a los diferentes paths (path, letra)
// var letraDelPath = make(map[string]string)

// // Índice para la siguiente letra disponible en el abecedario
// var nextLetterIndex = 0

// particiones montadaa
// var (
// 	// almacena particiones montada y el path del disco al que pertenece
// 	ParticionesMontadas map[string]string = make(map[string]string)
// 	// almacena información de los discos MOntados
// 	Montaje map[string]Montadas = make(map[string]Montadas)
// )

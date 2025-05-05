package global

// literalmente almacena la información del usuario loggeado actualmente
type Sesion struct {
	IsLoggedIn  bool
	Username    string
	Password    string
	PartitionID string
}

// para ingresar información al auth
func (usuario *Sesion) Login(username, password, partitionID string) {
	usuario.IsLoggedIn = true
	usuario.Username = username
	usuario.Password = password
	usuario.PartitionID = partitionID
}

// borra la información y resetea el usuarioactual
func (usuario *Sesion) Logout() {
	usuario.IsLoggedIn = false
	usuario.Username = ""
	usuario.Password = ""
	usuario.PartitionID = ""
}

// devuelve si hay sesión iniciada actualmente
func (usuario *Sesion) HaySesionIniciada() bool {
	return usuario.IsLoggedIn
}

// devuelve el usuario, contraseña y el idParticion
func (usuario *Sesion) ObtenerUsuario() (string, string, string) {
	return usuario.Username, usuario.Password, usuario.PartitionID
}

// // instancia de usuario actual
// var usuarioActual = &Sesion{
// 	IsLoggedIn:  false,
// 	Username:    "",
// 	Password:    "",
// 	PartitionID: "",
// }

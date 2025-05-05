
// interfaz para el usuario
interface UserContextInterface {
    isLoggedIn: boolean;
    name: string;
    password: string;
    idPart: string;
    diskName: string,
    login: (name: string, pwd: string, id: string) => void;
    logout: () => void;
    setDiskName: (nombreDisco: string) => void;
}

export default UserContextInterface
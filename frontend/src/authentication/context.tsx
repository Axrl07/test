import { createContext } from "react";
import UserContextInterface from "./interface";

// Creamos un contexto con valores por defecto (solo para tipado)
const UserContext = createContext<UserContextInterface>({
    isLoggedIn: false,
    name: '',
    password: '',
    idPart: '',
    diskName: '',
    login: () => {},
    logout: () => {},
    setDiskName: () => {},
});

export default UserContext;
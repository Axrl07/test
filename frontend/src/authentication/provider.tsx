import React, { useState } from "react";
import UserContext from "./context";
import UserContextInterface from "./interface";

interface Props {
    children: React.ReactNode;
}

const UserProvider: React.FC<Props> = ({ children }) => {
    // Estado del usuario sin incluir las funciones login/logout
    const [user, setUser] = useState<Omit<UserContextInterface, 'login' | 'logout' | 'setDiskName'>>({
        isLoggedIn: false,
        name: '',
        password: '',
        idPart: '',
        diskName: '',
    });

    const login = (name: string, pwd: string, id: string) => {
        setUser({
            isLoggedIn: true,
            name,
            password: pwd,
            idPart: id,
            diskName: '',
        });
    };

    const setDiskName = (nombreDisco: string) => {
        setUser(prev => ({
            ...prev,
            diskName: nombreDisco,
        }));
    };
    

    const logout = () => {
        setUser({
            isLoggedIn: false,
            name: '',
            password: '',
            idPart: '',
            diskName: '',
        });
    };

    return (
        <UserContext.Provider value={{ ...user, login, logout, setDiskName }}>
            {children}
        </UserContext.Provider>
    );
};

export default UserProvider;

// export default Menu;
import React, { useContext } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import UserContext from '../authentication/context';
import { CerrarSesion, VerificandoSesionActual } from '../services/api';

const Menu: React.FC = () => {
    const { isLoggedIn, name, password, idPart, login, logout } = useContext(UserContext);
    const leyenda = isLoggedIn ? `Bienvenido ${name}` : "No hay sesión iniciada";
    const navigate = useNavigate();

    const handleOpenNavbar = async () => {
        try {
            const usuario = await VerificandoSesionActual();
            login(usuario.nombre, usuario.clave, usuario.idParticion);
        } catch (error) {
            console.log("No hay sesión activa " + error);
            logout()
        }
    };

    const handlerLogout = async () => {
        try {
            const respuesta = await CerrarSesion()
            logout()
            alert(respuesta)
            navigate("/")
        } catch (err) {
            alert(err)
        }
    };

    return (
        <nav className="navbar bg-body-tertiary" data-bs-theme="dark">
            <div className="container-fluid">
                <span className="navbar-brand mb-0 h1">{leyenda}</span>

                <button
                    className="navbar-toggler"
                    type="button"
                    data-bs-toggle="offcanvas"
                    data-bs-target="#offcanvasNavbar"
                    aria-controls="offcanvasNavbar"
                    aria-label="Toggle navigation"
                    onClick={handleOpenNavbar} // ← Solo cuando haces click
                >
                    <span className="navbar-toggler-icon"></span>
                </button>

                <div
                    className="offcanvas offcanvas-end"
                    tabIndex={-1}
                    id="offcanvasNavbar"
                    aria-labelledby="offcanvasNavbarLabel"
                >
                    <div className="offcanvas-header">
                        <h5 className="offcanvas-title" id="offcanvasNavbarLabel">Panel de usuario</h5>
                        <button
                            type="button"
                            className="btn-close"
                            data-bs-dismiss="offcanvas"
                            aria-label="Close"
                        ></button>
                    </div>

                    <div className="offcanvas-body">
                        <ul className="navbar-nav justify-content-end flex-grow-1 pe-3">
                            <li className="nav-item">
                                <Link className="nav-link" to="/">Inicio</Link>
                            </li>
                            {isLoggedIn && (
                                <>
                                    <li className="nav-item">
                                        <Link className="nav-link" to="/discos">Explorador de archivos</Link>
                                    </li>
                                    <li className="nav-item">
                                        <Link className="nav-link" to="/journal">Tabla Journaling</Link>
                                    </li>
                                    <li><hr className="divider" /></li>
                                    <li className="nav-item">
                                        <p className="mb-1">Datos del usuario actual:</p>
                                        <p className="mb-1">Nombre: {name}</p>
                                        <p className="mb-1">Clave: {password}</p>
                                        <p className="mb-1">ID Partición: {idPart}</p>
                                    </li>
                                    <li><hr className="divider" /></li>
                                </>
                            )}



                            <li className="nav-item">
                                {isLoggedIn ? (
                                    <button className="btn btn-outline-danger w-100 mt-2" onClick={handlerLogout}>
                                        Cerrar Sesión
                                    </button>
                                ) : (
                                    <>
                                        <Link to="/login" className="btn btn-outline-success w-100 mt-2">
                                            Iniciar Sesión
                                        </Link>
                                    </>
                                )}
                            </li>
                        </ul>
                    </div>
                </div>
            </div>
        </nav>
    );
};

export default Menu;

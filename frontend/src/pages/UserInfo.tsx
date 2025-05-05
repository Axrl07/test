import { useNavigate } from 'react-router-dom';
import styles from '../styles/Login.module.css';
import { IniciarSesion } from '../services/api';
import React, { useContext } from 'react';
import UserContext from '../authentication/context';
const UserInfo: React.FC = () => {
  const { login } = useContext(UserContext);
  const navigate = useNavigate();

  const handleRegresar = () => {
    navigate("/");
  };

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;

    const nombre = (form.nombreUsuario as HTMLInputElement).value;
    const clave = (form.clave as HTMLInputElement).value;
    const idPart = (form.particionId as HTMLInputElement).value;

    // aquí si vamos a logearnos 
    try{
      const respuesta = await IniciarSesion(nombre, clave, idPart)
      login(nombre, clave, idPart) //actualizamos con los datos locales pues coincidieron
      alert(respuesta)
      navigate("/")
    }catch(err){
      alert(err)
    }
  };

  return (
    <form onSubmit={handleLogin} className={styles.formContainer}>
      <div>
        <h1 className={styles.h1}>Inicio de sesión</h1>
      </div>

      <div className="input-group mb-3">
        <span className="input-group-text">Usuario</span>
        <input
          type="text"
          className="form-control"
          placeholder="Ingrese su nombre de usuario"
          name="nombreUsuario"
          required
        />
      </div>

      <div className="input-group mb-3">
        <span className="input-group-text">Contraseña</span>
        <input
          type="password"
          className="form-control"
          placeholder="Ingrese su contraseña"
          name="clave"
          required
        />
      </div>

      <div className="input-group mb-3">
        <span className="input-group-text">ID Partición</span>
        <input
          type="text"
          className="form-control"
          placeholder="Ingrese el Id de la partición"
          name="particionId"
          required
        />
      </div>

      <div className="d-grid gap-2">
        <button type="submit" className="btn btn-success">Inicio de Sesión</button>
        <button type="button" onClick={handleRegresar} className="btn btn-warning">Regresar</button>
      </div>
    </form>
  );
};

export default UserInfo;


import fileIcon from "../assets/file-icon.png";
import { ArchivoInterface } from "../services/api";

interface FileProps {
    file: ArchivoInterface;
    onClick: () => void;
}

const Archivo: React.FC<FileProps> = ({ file, onClick }) => {
    return (
        <div
            className="card text-center p-3 border-0 shadow-sm"
            style={{ width: "140px", cursor: "pointer" }}
            onClick={onClick}
        >
            <img src={fileIcon} alt="File Icon" className="card-img-top mx-auto" style={{ width: "80px" }} />
            <div className="card-body p-1">
                <p className="card-text small fw-semibold">{file.nombre}</p>
            </div>
        </div>
    );
};

export default Archivo;

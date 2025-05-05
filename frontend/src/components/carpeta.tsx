import folderIcon from "../assets/folder-icon.png";
import { CarpetaInterface } from "../services/api";

interface FolderProps {
    folder: CarpetaInterface;
    onClick: () => void;
}

const Carpeta: React.FC<FolderProps> = ({ folder, onClick }) => {
    return (
        <div
            className="card text-center p-3 border-0 shadow-sm"
            style={{ width: "140px", cursor: "pointer" }}
            onClick={onClick}
        >
            <img src={folderIcon} alt="Folder Icon" className="card-img-top mx-auto" style={{ width: "80px" }} />
            <div className="card-body p-1">
                <p className="card-text small fw-semibold">{folder.nombre}</p>
            </div>
        </div>
    );
};

export default Carpeta;

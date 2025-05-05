import React, { useRef, useState, Suspense } from 'react'
const MonacoEditor = React.lazy(() => import('@monaco-editor/react'))
import { interpretarComandos } from '../services/api'

const Home: React.FC = () => {
    const [editor, setEditor] = useState('')
    const [consola, setConsola] = useState('')
    const inputRef = useRef<HTMLInputElement>(null)

    const interpretar = async () => {
        if (!editor.trim()) {
            setConsola("Ingrese al menos un comando para solicitar el analisis.")
        } else {
            const contenido = await interpretarComandos(editor.trim())
            setConsola(contenido)
        }
    }

    const handleClick = () => {
        inputRef.current?.click()
    }

    const cargar = (event: React.ChangeEvent<HTMLInputElement>) => {
        const archivo = event.target.files?.[0]
        if (archivo && archivo.name.endsWith(".smia")) {
            const reader = new FileReader()
            reader.onload = (e) => {
                setEditor(e.target?.result as string)
            }
            reader.readAsText(archivo)
        } else {
            alert("Por favor selecciona un archivo con extensión .smia")
        }
    }

    const limpiar = () => {
        if (inputRef.current) {
            inputRef.current.value = ""
        }
        setEditor('')
        setConsola('')
    }

    const guardar = () => {
        const blob = new Blob([editor], { type: 'text/plain' })
        const enlace = document.createElement('a')
        enlace.href = URL.createObjectURL(blob)
        enlace.download = 'archivo.smia'
        enlace.click()
    }

    return (
        <div className="container mt-4">
            {/* Botones arriba a la derecha */}
            <div className="d-flex justify-content-end gap-2 mb-3">
                <button className="btn btn-primary" onClick={interpretar}>Interpretar</button>
                <button className="btn btn-secondary" onClick={handleClick}>Cargar Archivo</button>
                <input
                    type="file"
                    accept=".smia"
                    style={{ display: "none" }}
                    ref={inputRef}
                    onChange={cargar}
                />
                <button className="btn btn-success" onClick={guardar}>Guardar archivo</button>
                <button className="btn btn-danger" onClick={limpiar}>Limpiar</button>
            </div>

            {/* Editores debajo de los botones */}
            <div className="row">
                <div className="col-12 mb-4">
                    <Suspense fallback={<div>Cargando editor...</div>}>
                        <MonacoEditor
                            height="40vh"
                            width="100%"
                            defaultLanguage="go"
                            theme="vs-dark"
                            value={editor}
                            onChange={(value) => setEditor(value as string)}
                        />
                    </Suspense>
                </div>
                <div className="col-12">
                    <Suspense fallback={<div>Cargando consola...</div>}>
                        <MonacoEditor
                            height="40vh"
                            width="100%"
                            defaultLanguage="go"
                            value={consola}
                            theme="vs-dark"
                            options={{ readOnly: true }}
                        />
                    </Suspense>
                </div>
            </div>
        </div>
    )
}

export default Home

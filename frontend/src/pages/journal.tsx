import React, { useState, useEffect } from 'react'
import { ObtenerTablaJournaling } from '../services/api'

const Journal: React.FC = () => {
    const [tabla, setTabla] = useState("")

    const handlerTabla = async () => {
        try {
            const respuesta = await ObtenerTablaJournaling()
            setTabla(respuesta)
        } catch (err) {
            alert(err)
        }
    }

    useEffect(() => {
        handlerTabla()
    }, [])

    return (
        <div className="container mt-4">
            <h1>Reporte Journaling</h1>
            <div className="row" dangerouslySetInnerHTML={{ __html: tabla }} />
        </div>
    )
}

export default Journal


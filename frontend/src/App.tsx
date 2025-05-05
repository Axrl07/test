import { BrowserRouter, Route } from 'react-router'
import Home from './pages/home'
import Menu from './components/menu'
import UserInfo from './pages/UserInfo'
import { Routes } from 'react-router'
import "./styles/App.css"
import Journal from './pages/journal'
import Discos from './pages/discos'
import Particiones from './pages/particiones'
import ExploradorArchivos from './pages/explorador'

function App() {

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<><Menu /><Home /></>} />
        <Route path="/login" element={<><Menu /><UserInfo /></>} />
        <Route path="/journal" element={<><Menu /><Journal /></>} />
        <Route path="/discos" element={<><Menu /><Discos /></>} />
        <Route path="/:nombreDisco/particiones" element={<><Menu /><Particiones /></>} />
        <Route path="/:idPart/sistema" element={<><Menu /><ExploradorArchivos /></>} />
      </Routes>
    </BrowserRouter>
  )
}

export default App

import { Link } from 'react-router-dom'

function Home() {
  return (
    <div>
      <h1>Eva Trainer</h1>
      <nav>
        <Link to="/editor">Editeur de zones</Link>
      </nav>
    </div>
  )
}

export default Home

import React, {Component} from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'

// ErrorBoundary raíz: si algo rompe el render de la app, en lugar de una
// pantalla en blanco se muestra este fallback con la opción de recargar.
class RootBoundary extends Component<{children: React.ReactNode}, {error: Error | null}> {
    constructor(props: {children: React.ReactNode}) {
        super(props);
        this.state = {error: null};
    }

    static getDerivedStateFromError(error: Error) {
        return {error};
    }

    componentDidCatch(error: Error) {
        console.error("RootBoundary:", error);
    }

    render() {
        if (this.state.error) {
            return (
                <div id="root-error">
                    <h2>Ups, algo salió mal</h2>
                    <p>La interfaz se detuvo por un error inesperado.</p>
                    <pre>{this.state.error.message}</pre>
                    <button type="button" className="btn primary" onClick={() => window.location.reload()}>
                        Recargar interfaz
                    </button>
                </div>
            );
        }
        return this.props.children;
    }
}

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <RootBoundary>
            <App/>
        </RootBoundary>
    </React.StrictMode>
)
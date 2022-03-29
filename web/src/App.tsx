import React from 'react';
import { Routes,
  Route,
  useNavigate,
  useLocation,
  Navigate,
} from "react-router-dom";
import DefaultLayout from "./pages/layout/DefaultLayout";
import LoginLayout from "./pages/layout/LoginLayout";
import TablePage from "./pages/TablePage";
import LoginPage from "./pages/login/LoginPage";
import './App.scss';


function App() {
  return (
    <div className="App torq">
      <AuthProvider>
        <Routes>
          <Route element={<LoginLayout/>}>
            <Route path="/login" element={<LoginPage/>}/>
          </Route>
          <Route element={<DefaultLayout/>}>
            <Route path="/" element={<TablePage/>} />
          </Route>
        </Routes>
      </AuthProvider>
    </div>
  );
}




interface AuthContextType {
  user: any;
  signin: (user: string, callback: VoidFunction) => void;
  signout: (callback: VoidFunction) => void;
}

let AuthContext = React.createContext<AuthContextType>(null!);

function useAuth() {
  return React.useContext(AuthContext);
}

const fakeAuthProvider = {
  isAuthenticated: false,
  signin(callback: VoidFunction) {
    fakeAuthProvider.isAuthenticated = true;
    setTimeout(callback, 100); // fake async
  },
  signout(callback: VoidFunction) {
    fakeAuthProvider.isAuthenticated = false;
    setTimeout(callback, 100);
  },
};

function AuthProvider({ children }: { children: React.ReactNode }) {
  let [user, setUser] = React.useState<any>(null);

  let signin = (newUser: string, callback: VoidFunction) => {
    return fakeAuthProvider.signin(() => {
      setUser(newUser);
      callback();
    });
  };

  let signout = (callback: VoidFunction) => {
    return fakeAuthProvider.signout(() => {
      setUser(null);
      callback();
    });
  };

  let value = { user, signin, signout };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

function RequireAuth({ children }: { children: JSX.Element }) {
  let auth = useAuth();
  let location = useLocation();

  if (!auth.user) {
    // Redirect them to the /login page, but save the current location they were
    // trying to go to when they were redirected. This allows us to send them
    // along to that page after they login, which is a nicer user experience
    // than dropping them off on the home page.
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
}

export default App;

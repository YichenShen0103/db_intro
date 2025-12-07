import { useEffect, useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { userAPI } from '../api'

function Layout() {
    const navigate = useNavigate()
    const location = useLocation()
    const currentPath = location.pathname
    const [username, setUsername] = useState('')

    useEffect(() => {
        const fetchProfile = async () => {
            try {
                const res = await userAPI.getProfile()
                setUsername(res.data?.username || '')
            } catch (err) {
                console.error('获取用户信息失败：', err)
            }
        }

        fetchProfile()
    }, [])

    const handleLogout = () => {
        localStorage.removeItem('token')
        navigate('/login')
    }

    return (
        <div className="flex h-screen bg-gray-50">
            {/* Sidebar */}
            <aside className="w-64 bg-white shadow-md flex flex-col">
                <div className="p-6 border-b">
                    <h1 className="text-xl font-bold text-blue-600">
                        <i className="fas fa-chart-pie mr-2"></i>科研数据汇总
                    </h1>
                </div>
                <nav className="flex-1 p-4 space-y-2">
                    <button
                        onClick={() => navigate('/projects')}
                        className={`w-full text-left block px-4 py-2 rounded hover:bg-gray-100 transition ${currentPath.startsWith('/projects') ? 'bg-blue-50 text-blue-600' : ''
                            }`}
                    >
                        <i className="fas fa-folder-open mr-2"></i> 项目管理
                    </button>
                    <button
                        onClick={() => navigate('/teachers')}
                        className={`w-full text-left block px-4 py-2 rounded hover:bg-gray-100 transition ${currentPath.startsWith('/teachers') ? 'bg-blue-50 text-blue-600' : ''
                            }`}
                    >
                        <i className="fas fa-users mr-2"></i> 教师信息库
                    </button>
                    <button
                        onClick={() => navigate('/settings')}
                        className={`w-full text-left block px-4 py-2 rounded hover:bg-gray-100 transition ${currentPath.startsWith('/settings') ? 'bg-blue-50 text-blue-600' : ''
                            }`}
                    >
                        <i className="fas fa-cog mr-2"></i> 邮箱设置
                    </button>
                </nav>
                <div className="p-4 border-t text-xs text-gray-400 space-y-2">
                    <div className="flex items-center justify-between gap-3">
                        <div className="flex items-center text-sm text-gray-600 truncate">
                            <i className="fas fa-user-circle mr-2 text-blue-500"></i>
                            {username || '加载中...'}
                        </div>
                        <button
                            onClick={handleLogout}
                            className="px-3 py-2 text-sm text-red-600 border border-red-200 rounded hover:bg-red-50 transition whitespace-nowrap"
                        >
                            <i className="fas fa-sign-out-alt mr-1"></i> 退出登录
                        </button>
                    </div>
                    <div>version 1.0</div>
                </div>
            </aside>

            {/* Main Content */}
            <main className="flex-1 overflow-auto p-8">
                <Outlet />
            </main>
        </div>
    )
}

export default Layout

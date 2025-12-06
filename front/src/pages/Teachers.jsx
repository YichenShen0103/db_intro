import { useState, useEffect } from 'react'
import { teachersAPI, departmentsAPI } from '../api'

function Teachers() {
    const [teachers, setTeachers] = useState([])
    const [departments, setDepartments] = useState([])
    const [showCreateTeacherModal, setShowCreateTeacherModal] = useState(false)
    const [teacherForm, setTeacherForm] = useState({
        name: '',
        email: '',
        department_id: '',
        phone: '',
    })

    useEffect(() => {
        loadTeachers()
        loadDepartments()
    }, [])

    const loadTeachers = async () => {
        try {
            const res = await teachersAPI.getAll()
            setTeachers(res.data?.data || [])
        } catch (err) {
            console.error('Failed to load teachers:', err)
        }
    }

    const loadDepartments = async () => {
        try {
            const res = await departmentsAPI.getAll()
            setDepartments(res.data?.data || [])
        } catch (err) {
            console.error('Failed to load departments:', err)
        }
    }

    const createTeacher = async () => {
        try {
            const data = {
                ...teacherForm,
                department_id: teacherForm.department_id ? parseInt(teacherForm.department_id) : null
            }
            await teachersAPI.create(data)
            setShowCreateTeacherModal(false)
            setTeacherForm({ name: '', email: '', department_id: '', phone: '' })
            loadTeachers()
            alert('教师添加成功！')
        } catch (err) {
            alert('添加失败：' + (err.response?.data?.error || err.message))
        }
    }

    return (
        <div>
            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold">教师信息库</h2>
                <button
                    onClick={() => setShowCreateTeacherModal(true)}
                    className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition"
                >
                    <i className="fas fa-user-plus mr-1"></i> 添加教师
                </button>
            </div>
            <div className="bg-white shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
                <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                        <tr>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                                姓名
                            </th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                                系别
                            </th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                                邮箱
                            </th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                                电话
                            </th>
                            <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                                操作
                            </th>
                        </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                        {teachers.map((teacher) => (
                            <tr key={teacher.id}>
                                <td className="px-6 py-4 whitespace-nowrap font-medium text-gray-900">
                                    {teacher.name}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-gray-500">
                                    {teacher.department_name || '-'}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-gray-500">{teacher.email}</td>
                                <td className="px-6 py-4 whitespace-nowrap text-gray-500">{teacher.phone || '-'}</td>
                                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                    <button className="text-indigo-600 hover:text-indigo-900">编辑</button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {/* Create Teacher Modal */}
            {showCreateTeacherModal && (
                <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center z-50">
                    <div className="bg-white p-8 rounded-lg shadow-xl w-1/3">
                        <h3 className="text-xl font-bold mb-4">添加教师</h3>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700">姓名</label>
                                <input
                                    type="text"
                                    value={teacherForm.name}
                                    onChange={(e) => setTeacherForm({ ...teacherForm, name: e.target.value })}
                                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700">邮箱</label>
                                <input
                                    type="email"
                                    value={teacherForm.email}
                                    onChange={(e) => setTeacherForm({ ...teacherForm, email: e.target.value })}
                                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700">所在系</label>
                                <select
                                    value={teacherForm.department_id}
                                    onChange={(e) => setTeacherForm({ ...teacherForm, department_id: e.target.value })}
                                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                                >
                                    <option value="">选择系别</option>
                                    {departments.map((dept) => (
                                        <option key={dept.id} value={dept.id}>
                                            {dept.name}
                                        </option>
                                    ))}
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700">电话</label>
                                <input
                                    type="text"
                                    value={teacherForm.phone}
                                    onChange={(e) => setTeacherForm({ ...teacherForm, phone: e.target.value })}
                                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                                />
                            </div>
                        </div>
                        <div className="mt-6 flex justify-end space-x-3">
                            <button
                                onClick={() => setShowCreateTeacherModal(false)}
                                className="bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300"
                            >
                                取消
                            </button>
                            <button
                                onClick={createTeacher}
                                className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                            >
                                添加
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}

export default Teachers

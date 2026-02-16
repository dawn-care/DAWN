import { useNavigate, useParams } from 'react-router-dom';
import { useState, useEffect } from 'react';
import AgendaList, { type ConfigurationWithMetadata } from '../components/AgendaList';
import Library from './Library';

export default function AgendaView() {
    const { date } = useParams();
    const navigate = useNavigate();
    const [items, setItems] = useState<ConfigurationWithMetadata[]>([]);
    const [isLibraryModalOpen, setIsLibraryModalOpen] = useState(false);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                setLoading(true);
                setError(null);

                const { dayOfMonth, month, dayOfWeek } = currentDateToComponents();
                if (dayOfMonth === 0) return;

                const params = new URLSearchParams({
                    carepartner_id: '1',
                    day: String(dayOfMonth),
                    month: String(month),
                    day_of_week: String(dayOfWeek),
                });

                const res = await fetch(`http://localhost:8080/configurations/day?${params}`);
                if (!res.ok) throw new Error('Failed to fetch configurations');
                const data: ConfigurationWithMetadata[] = await res.json();
                setItems(data || []);
            } catch (err) {
                const errorMessage = err instanceof Error ? err.message : 'An error occurred';
                setError(errorMessage);
                console.error('Error fetching data:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, [date]);

    const currentDateToComponents = () => {
        if (!date) return { dayOfMonth: 0, month: 0, dayOfWeek: 0 };
        const currentDate = new Date(date);
        const dayOfWeek = currentDate.getDay(); // 0 (Sun) to 6 (Sat)
        const dayOfMonth = currentDate.getDate(); // 1 to 31
        const month = currentDate.getMonth() + 1;
        return { dayOfMonth, month, dayOfWeek };
    }

    const handlePreviousDay = () => {
        if (!date) return;
        const currentDate = new Date(date);
        currentDate.setDate(currentDate.getDate() - 1);
        const newDate = currentDate.toISOString().split('T')[0]; // Format: YYYY-MM-DD
        navigate(`/agenda/${newDate}`);
    };

    const handleNextDay = () => {
        if (!date) return;
        const currentDate = new Date(date);
        currentDate.setDate(currentDate.getDate() + 1);
        const newDate = currentDate.toISOString().split('T')[0]; // Format: YYYY-MM-DD
        navigate(`/agenda/${newDate}`);
    };

    return (
        <div className="min-h-screen bg-white">
            {loading && (
                <div className="flex items-center justify-center min-h-screen">
                    <p className="text-xl text-gray-600">Loading agenda...</p>
                </div>
            )}
            
            {error && (
                <div className="flex items-center justify-center min-h-screen">
                    <div className="bg-red-50 border-2 border-red-400 rounded-lg p-6">
                        <p className="text-red-700 font-semibold">Error loading data</p>
                        <p className="text-red-600 text-sm mt-2">{error}</p>
                    </div>
                </div>
            )}
            
            {!loading && !error && (
            <div className="px-8 py-8">
                <div className="flex items-center justify-between mb-8">
                    <button onClick={handlePreviousDay} className="text-4xl text-gray-800 hover:text-gray-600">&lt;</button>
                    <h2 className="text-4xl font-light text-gray-800">{date} Videos for Dad</h2>
                    <button onClick={handleNextDay} className="text-4xl text-gray-800 hover:text-gray-600">&gt;</button>
                </div>

                <div className="flex gap-8 mb-8">
                    <div className="flex items-center gap-4">
                        <label className="text-gray-700 font-medium">View:</label>
                        <button className="px-6 py-2 bg-gray-200 border-b-2 border-gray-400 rounded text-gray-800 flex items-center gap-2 hover:bg-gray-300">
                            Daily
                            <span>∨</span>
                        </button>
                    </div>
                    <div className="flex items-center gap-4">
                        <label className="text-gray-700 font-medium">Care Partner:</label>
                        <button className="px-6 py-2 bg-gray-200 border-b-2 border-gray-400 rounded text-gray-800 flex items-center gap-2 hover:bg-gray-300">
                            Dad
                            <span>∨</span>
                        </button>
                    </div>
                </div>

                <AgendaList items={items} setItems={setItems} />

                <div className="flex items-center justify-between mt-12">
                    <button className="text-3xl text-gray-600 hover:text-gray-800">⚙</button>
                    <div className="flex gap-6">
                        <button className="px-8 py-3 border-2 border-cyan-400 text-cyan-400 rounded-lg font-semibold hover:bg-cyan-50">
                            Preview
                        </button>
                        <button onClick={() => setIsLibraryModalOpen(true)} className="px-8 py-3 border-2 border-cyan-400 bg-white text-cyan-400 rounded-lg text-3xl font-bold hover:bg-cyan-50">
                            +
                        </button>
                    </div>
                </div>

                {isLibraryModalOpen && (
                    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                        <div className="bg-white rounded-lg shadow-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto relative">
                            <button
                                onClick={() => setIsLibraryModalOpen(false)}
                                className="absolute top-4 right-4 text-gray-600 hover:text-gray-800 text-2xl z-10"
                            >
                                ✕
                            </button>
                            <div className="p-8">
                                <Library />
                            </div>
                        </div>
                    </div>
                )}
            </div>
            )}
        </div>
    );
}
import { useNavigate, useParams } from 'react-router-dom';
import { useState, useEffect } from 'react';
import AgendaList, { type ConfigurationItem, type Module, type ModuleMetadata }  from '../components/AgendaList';
import Library from './Library';

export default function AgendaView() {
    const { date } = useParams();
    const navigate = useNavigate();
    const [items, setItems] = useState<ConfigurationItem[]>([]);
    const [isLibraryModalOpen, setIsLibraryModalOpen] = useState(false);
    const [modules, setModules] = useState<Module[]>([]);
    const [metadata, setMetadata] = useState<Record<number, ModuleMetadata>>({});
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                setLoading(true);
                setError(null);

                // Fetch configurations
                const { dayOfMonth, month, dayOfWeek } = currentDateToComponents();
                const configRes = await fetch(`http://localhost:8080/configurations?carepartner_id=1&day=${dayOfMonth}&month=${month}&dayOfWeek=${dayOfWeek}`);
                if (!configRes.ok) throw new Error('Failed to fetch configurations');
                const configData = await configRes.json();
                setItems(configData || []);

                // Fetch modules
                const modulesRes = await fetch('http://localhost:8080/api/modules?carepartner_id=1');
                if (!modulesRes.ok) throw new Error('Failed to fetch modules');
                const modulesData = await modulesRes.json();
                setModules(modulesData || []);

                // Fetch metadata for each module
                const metadataMap: Record<number, ModuleMetadata> = {};
                
                for (const module of modulesData || []) {
                    try {
                        let metaRes;
                        if (module.module_type === 'media') {
                            metaRes = await fetch(`http://localhost:8080/api/media/${module.metadata_id}`);
                        } else if (module.module_type === 'widget') {
                            metaRes = await fetch(`http://localhost:8080/api/widgets/${module.metadata_id}`);
                        } else if (module.module_type === 'prompt') {
                            metaRes = await fetch(`http://localhost:8080/api/prompts/${module.metadata_id}`);
                        }
                        
                        if (metaRes && metaRes.ok) {
                            const metaData = await metaRes.json();
                            metadataMap[module.metadata_id] = metaData;
                        }
                    } catch (err) {
                        console.warn(`Failed to fetch metadata for module ${module.metadata_id}:`, err);
                    }
                }
                
                setMetadata(metadataMap);
            } catch (err) {
                const errorMessage = err instanceof Error ? err.message : 'An error occurred';
                setError(errorMessage);
                console.error('Error fetching data:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

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

                <AgendaList items={items} setItems={setItems} modules={modules} setModules={setModules} metadata={metadata} setMetadata={setMetadata} />

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
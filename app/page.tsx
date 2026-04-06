export default function Home() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-500 to-purple-600">
      <div className="text-center text-white p-8">
        <h1 className="text-5xl font-bold mb-4">WMS</h1>
        <p className="text-xl mb-8">Warehouse Management System</p>
        <a
          href="/login"
          className="inline-block bg-white text-blue-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-100 transition"
        >
          Get Started
        </a>
      </div>
    </div>
  );
}

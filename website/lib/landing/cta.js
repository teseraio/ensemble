/* This example requires Tailwind CSS v2.0+ */
export default function Example() {
  return (
    <div className="bg-black">
      <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 md:py-16 lg:px-8 lg:py-20">
        <div className="border border-white p-10">
          <h2 className="text-3xl font-extrabold tracking-tight text-white sm:text-4xl">
            <span className="block">Ready to dive in?</span>
            <span className="block">Start your free trial today.</span>
          </h2>
          <div className="mt-8 flex">
            <div className="inline-flex rounded-md shadow">
              <a
                href="#"
                className="inline-flex items-center justify-center px-5 py-3 border border-transparent text-base font-medium bg-ensemble hover:bg-white hover:text-ensemble"
              >
                Learn more
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

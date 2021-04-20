
export default function Changelog({vers}) {
    return (
        <div>
        <div class="bg-white pt-10 pb-20 px-4 sm:px-6 lg:pt-24 lg:pb-28 lg:px-8">
          <div class="relative max-w-4xl mx-auto divide-y-2 divide-gray-200 lg:max-w-7xl">
            <div>
              <h2 class="text-3xl tracking-tight font-extrabold text-gray-900 sm:text-4xl">
                Changelog
              </h2>
            </div>
            <div class="mt-6 pt-10 grid gap-16">

              {vers.map((item, indx) => (
                <div className="grid grid-cols-5">
                    <div className="hidden lg:block">
                        {item.version}
                    </div>
                    <div className="col-span-5 lg:col-span-4 text-base">
                        <div dangerouslySetInnerHTML={{__html: item.content}} />
                    </div>
                </div>
              ))}
        
            </div>
          </div>
        </div>
    </div>
    )
}
